// Package repository содержит реализацию хранилища метрик.
// Внутреннее хранилище MemStorage позволяет сохранять
// значения типа Gauge и Counter, а также обновлять их.
package repository

import models "github.com/F3dosik/metalert.git/internal/model"

type Storage interface {
	UpdateGauge(name string, value models.Gauge)
	UpdateCounter(name string, value models.Counter)
}

type MemStorage struct {
	Gauges   map[string]models.Gauge
	Counters map[string]models.Counter
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}
}

func (mS *MemStorage) UpdateGauge(name string, value models.Gauge) {
	mS.Gauges[name] = value
}

func (mS *MemStorage) UpdateCounter(name string, value models.Counter) {
	mS.Counters[name] += value
}
