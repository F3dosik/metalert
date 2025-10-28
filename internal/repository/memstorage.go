// Package repository содержит реализацию хранилища метрик.
// Внутреннее хранилище MemStorage позволяет сохранять
// значения типа Gauge и Counter, а также обновлять их.
package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/F3dosik/metalert.git/pkg/models"
)

type MemMetricsStorage struct {
	Gauges   map[string]models.Gauge
	Counters map[string]models.Counter
	mutex    sync.RWMutex
}

func NewMemMetricsStorage() *MemMetricsStorage {
	return &MemMetricsStorage{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}
}

func (f *MemMetricsStorage) SetGauge(ctx context.Context, metName string, value models.Gauge) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.Gauges[metName] = value
	return nil
}

func (f *MemMetricsStorage) GetGauge(ctx context.Context, metName string) (models.Gauge, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	v, ok := f.Gauges[metName]
	if !ok {
		return 0, fmt.Errorf("метрика %s отсутствует", metName)
	}

	return v, nil
}

func (f *MemMetricsStorage) AddCounter(ctx context.Context, metName string, value models.Counter) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.Counters[metName] += value
	return nil
}

func (f *MemMetricsStorage) SetCounter(ctx context.Context, metName string, value models.Counter) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.Counters[metName] = value
	return nil
}

func (f *MemMetricsStorage) GetCounter(ctx context.Context, metName string) (models.Counter, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	v, ok := f.Counters[metName]
	if !ok {
		return 0, fmt.Errorf("метрика %s отсутствует", metName)
	}

	return v, nil
}

func (f *MemMetricsStorage) GetAllMetrics(ctx context.Context) ([]models.Metric, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	metrics := make([]models.Metric, 0, len(f.Gauges)+len(f.Counters))

	for name, value := range f.Gauges {
		metrics = append(metrics, models.Metric{
			ID:    name,
			MType: models.TypeGauge,
			Value: &value,
		})
	}

	for name, value := range f.Counters {
		metrics = append(metrics, models.Metric{
			ID:    name,
			MType: models.TypeCounter,
			Delta: &value,
		})
	}

	return metrics, nil
}
