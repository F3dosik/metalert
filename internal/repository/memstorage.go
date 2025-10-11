// Package repository содержит реализацию хранилища метрик.
// Внутреннее хранилище MemStorage позволяет сохранять
// значения типа Gauge и Counter, а также обновлять их.
package repository

import (
	"fmt"
	"sync"

	"github.com/F3dosik/metalert.git/pkg/models"
)

type MetricsStorage interface {
	SetGauge(name string, value models.Gauge)
	GetGauge(name string) (models.Gauge, error)

	AddCounter(name string, value models.Counter)
	GetCounter(name string) (models.Counter, error)

	GetAllMetrics() []models.Metric
}

type MemMetricsStorage struct {
	Gauges   map[string]models.Gauge
	Counters map[string]models.Counter
	mutex    sync.RWMutex
}

func NewMemMetricsStorage() *MemMetricsStorage {
	return &MemMetricsStorage{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
		mutex:    sync.RWMutex{},
	}
}

func (mS *MemMetricsStorage) SetGauge(metName string, value models.Gauge) {
	mS.mutex.Lock()
	defer mS.mutex.Unlock()

	mS.Gauges[metName] = value
}

func (mS *MemMetricsStorage) GetGauge(metName string) (models.Gauge, error) {
	mS.mutex.RLock()
	defer mS.mutex.RUnlock()

	value, exist := mS.Gauges[metName]
	if !exist {
		return 0, fmt.Errorf("метрика %s отсутствует", metName)
	}
	return value, nil
}

func (mS *MemMetricsStorage) UpdateCounter(metName string, value models.Counter) {
	mS.mutex.Lock()
	defer mS.mutex.Unlock()
	mS.Counters[metName] += value
}

func (mS *MemMetricsStorage) GetCounter(metName string) (models.Counter, error) {
	mS.mutex.RLock()
	defer mS.mutex.RUnlock()

	value, exist := mS.Counters[metName]
	if !exist {
		return 0, fmt.Errorf("метрика %s отсутствует", metName)
	}
	return value, nil
}

func (mS *MemMetricsStorage) GetAllMetrics() []models.Metric {
	mS.mutex.RLock()
	defer mS.mutex.RUnlock()

	metrics := make([]models.Metric, 0, len(mS.Gauges)+len(mS.Counters))

	for name, value := range mS.Gauges {
		metrics = append(metrics, models.Metric{
			ID:    name,
			MType: models.TypeGauge,
			Value: &value,
		})
	}

	for name, value := range mS.Counters {
		metrics = append(metrics, models.Metric{
			ID:    name,
			MType: models.TypeCounter,
			Delta: &value,
		})
	}
	return metrics

}
