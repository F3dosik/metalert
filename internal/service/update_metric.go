package service

import (
	"context"

	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/pkg/models"
)

func UpdateMetric(ctx context.Context, storage repository.MetricsStorage, metName string, metValue any) error {
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

	return nil
}

func UpdateMetricFromStruct(ctx context.Context, storage repository.MetricsStorage, met models.Metric) error {
	switch met.MType {
	case models.TypeGauge:
		if err := UpdateMetric(ctx, storage, met.ID, *met.Value); err != nil {
			return err
		}
	case models.TypeCounter:
		if err := UpdateMetric(ctx, storage, met.ID, *met.Delta); err != nil {
			return err
		}
	}

	return nil
}
