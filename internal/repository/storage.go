package repository

import "github.com/F3dosik/metalert.git/pkg/models"

type MetricsStorage interface {
	SetGauge(name string, value models.Gauge)
	GetGauge(name string) (models.Gauge, error)

	SetCounter(name string, value models.Counter)
	AddCounter(name string, value models.Counter)
	GetCounter(name string) (models.Counter, error)

	GetAllMetrics() []models.Metric
}

type Savable interface {
	Save() error
}
