package service

import (
	"context"
	"time"

	"github.com/F3dosik/metalert/internal/audit"
	"github.com/F3dosik/metalert/internal/repository"
	"github.com/F3dosik/metalert/pkg/models"
)

func UpdateMetric(ctx context.Context, storage repository.MetricsStorage, metName string, metValue any, dispatcher *audit.AuditDispatcher, ip string) error {
	switch v := metValue.(type) {
	case models.Gauge:
		if err := storage.SetGauge(ctx, metName, v); err != nil {
			return err
		}
	case models.Counter:
		if err := storage.AddCounter(ctx, metName, v); err != nil {
			return err
		}
	}

	if dispatcher != nil {
		dispatcher.Publish(audit.AuditEvent{
			Ts:        time.Now().Unix(),
			Metrics:   []string{metName},
			IPAddress: ip,
		})
	}

	return nil
}

func UpdateMetricFromStruct(ctx context.Context, storage repository.MetricsStorage, met models.Metric) error {
	switch met.MType {
	case models.TypeGauge:
		if err := UpdateMetric(ctx, storage, met.ID, *met.Value, nil, ""); err != nil {
			return err
		}
	case models.TypeCounter:
		if err := UpdateMetric(ctx, storage, met.ID, *met.Delta, nil, ""); err != nil {
			return err
		}
	}

	return nil
}

func UpdateMetrics(ctx context.Context, storage repository.MetricsStorage, metrics []models.Metric, dispatcher *audit.AuditDispatcher, ip string) error {
	for _, metric := range metrics {
		if err := ValidateMetric(metric); err != nil {
			return err
		}
	}

	if err := storage.UpdateMany(ctx, metrics); err != nil {
		return err
	}

	if dispatcher != nil {
		names := make([]string, len(metrics))
		for i, m := range metrics {
			names[i] = m.ID
		}
		dispatcher.Publish(audit.AuditEvent{
			Ts:        time.Now().Unix(),
			Metrics:   names,
			IPAddress: ip,
		})
	}

	return nil
}
