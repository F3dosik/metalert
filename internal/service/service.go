package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/F3dosik/metalert/internal/audit"
	"github.com/F3dosik/metalert/internal/repository"
	"github.com/F3dosik/metalert/pkg/models"
)

// ErrPingNotSupported is returned by Ping when the storage does not support it.
var ErrPingNotSupported = errors.New("storage does not support ping")

// MetricsService is the single entry point for all metric business operations.
type MetricsService interface {
	// Update parses rawValue, validates, and stores a single metric.
	Update(ctx context.Context, metType models.MetricType, name, rawValue, ip string) error

	// UpdateMany validates and stores a batch of metrics.
	UpdateMany(ctx context.Context, metrics []models.Metric, ip string) error

	// GetGauge returns the current value of a gauge metric.
	GetGauge(ctx context.Context, name string) (models.Gauge, error)

	// GetCounter returns the current value of a counter metric.
	GetCounter(ctx context.Context, name string) (models.Counter, error)

	// GetAll returns all stored metrics.
	GetAll(ctx context.Context) ([]models.Metric, error)

	// Ping checks the underlying storage connectivity.
	// Returns ErrPingNotSupported if the storage does not implement a Ping method.
	Ping() error
}

type pinger interface {
	Ping() error
}

type metricsService struct {
	storage      repository.MetricsStorage
	dispatcher   *audit.AuditDispatcher
	saveOnUpdate bool
	logger       *zap.SugaredLogger
}

// NewMetricsService constructs a MetricsService.
func NewMetricsService(
	storage repository.MetricsStorage,
	dispatcher *audit.AuditDispatcher,
	saveOnUpdate bool,
	logger *zap.SugaredLogger,
) MetricsService {
	return &metricsService{
		storage:      storage,
		dispatcher:   dispatcher,
		saveOnUpdate: saveOnUpdate,
		logger:       logger,
	}
}

func (s *metricsService) Update(ctx context.Context, metType models.MetricType, name, rawValue, ip string) error {
	value, err := CheckAndParseValue(metType, name, rawValue)
	if err != nil {
		return err
	}
	return UpdateMetric(ctx, s.storage, name, value, s.dispatcher, ip)
}

func (s *metricsService) UpdateMany(ctx context.Context, metrics []models.Metric, ip string) error {
	if err := UpdateMetrics(ctx, s.storage, metrics, s.dispatcher, ip); err != nil {
		return err
	}
	if s.saveOnUpdate {
		if savable, ok := s.storage.(repository.Savable); ok {
			go func() {
				if err := savable.Save(); err != nil {
					s.logger.Warnw("error saving metrics", "error", err)
				}
			}()
		}
	}
	return nil
}

func (s *metricsService) GetGauge(ctx context.Context, name string) (models.Gauge, error) {
	return s.storage.GetGauge(ctx, name)
}

func (s *metricsService) GetCounter(ctx context.Context, name string) (models.Counter, error) {
	return s.storage.GetCounter(ctx, name)
}

func (s *metricsService) GetAll(ctx context.Context) ([]models.Metric, error) {
	return s.storage.GetAllMetrics(ctx)
}

func (s *metricsService) Ping() error {
	p, ok := s.storage.(pinger)
	if !ok {
		return ErrPingNotSupported
	}
	return p.Ping()
}
