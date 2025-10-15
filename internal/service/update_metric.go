package service

import (
	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/pkg/models"
)

func UpdateMetric(storage *repository.MemMetricsStorage, metName string, metValue any) {
	switch v := metValue.(type) {
	case models.Gauge:
		storage.SetGauge(metName, v)
	case models.Counter:
		storage.AddCounter(metName, v)
	}
}

func UpdateMetricFromStruct(storage *repository.MemMetricsStorage, met models.Metric) {
	switch met.MType {
	case models.TypeGauge:
		UpdateMetric(storage, met.ID, *met.Value)
	case models.TypeCounter:
		UpdateMetric(storage, met.ID, *met.Delta)
	}
}
