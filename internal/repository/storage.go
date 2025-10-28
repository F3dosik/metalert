package repository

import (
	"context"

	"github.com/F3dosik/metalert.git/pkg/models"
)

type MetricsStorage interface {
	SetGauge(ctx context.Context, name string, value models.Gauge) error
	GetGauge(ctx context.Context, name string) (models.Gauge, error)

	AddCounter(ctx context.Context, name string, value models.Counter) error
	GetCounter(ctx context.Context, name string) (models.Counter, error)

	GetAllMetrics(ctx context.Context) ([]models.Metric, error)
}

type Savable interface {
	Save() error
}
