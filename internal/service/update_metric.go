package service

import (
	"context"

	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/pkg/models"
)

func UpdateMetric(ctx context.Context, storage repository.MetricsStorage, metName string, metValue any) {
	switch v := metValue.(type) {
	case models.Gauge:
		storage.SetGauge(ctx, metName, v)
	case models.Counter:
		storage.AddCounter(ctx, metName, v)
	}
}

func UpdateMetricFromStruct(ctx context.Context, storage repository.MetricsStorage, met models.Metric) {
	switch met.MType {
	case models.TypeGauge:
		UpdateMetric(ctx, storage, met.ID, *met.Value)
	case models.TypeCounter:
		UpdateMetric(ctx, storage, met.ID, *met.Delta)
	}
}
