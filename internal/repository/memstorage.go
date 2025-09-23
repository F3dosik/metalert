// Package repository содержит реализацию хранилища метрик.
// Внутреннее хранилище MemStorage позволяет сохранять
// значения типа Gauge и Counter, а также обновлять их.
package repository

import (
	"fmt"
	"sync"

	models "github.com/F3dosik/metalert.git/internal/model"
)

type Storage interface {
	SetGauge(name string, value models.Gauge)
	GetGauge(name string) (models.Gauge, error)

	UpdateCounter(name string, value models.Counter)
	GetCounter(name string) (models.Counter, error)

	GetAllMetrics() []models.Metrics
}

type MemStorage struct {
	Gauges   map[string]models.Gauge
	Counters map[string]models.Counter
	mutex    sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
		mutex:    sync.RWMutex{},
	}
}

func (mS *MemStorage) SetGauge(metName string, value models.Gauge) {
	mS.mutex.Lock()
	defer mS.mutex.Unlock()

	mS.Gauges[metName] = value
}

func (mS *MemStorage) GetGauge(metName string) (models.Gauge, error) {
	mS.mutex.RLock()
	defer mS.mutex.RUnlock()

	value, exist := mS.Gauges[metName]
	if !exist {
		return 0, fmt.Errorf("метрика %s отсутствует", metName)
	}
	return value, nil
}

func (mS *MemStorage) UpdateCounter(metName string, value models.Counter) {
	mS.mutex.Lock()
	defer mS.mutex.Unlock()

	mS.Counters[metName] += value
}

func (mS *MemStorage) GetCounter(metName string) (models.Counter, error) {
	mS.mutex.RLock()
	defer mS.mutex.RUnlock()

	value, exist := mS.Counters[metName]
	if !exist {
		return 0, fmt.Errorf("метрика %s отсутствует", metName)
	}
	return value, nil
}

func (mS *MemStorage) GetAllMetrics() []models.Metrics {
	mS.mutex.RLock()
	defer mS.mutex.RUnlock()

	metrics := make([]models.Metrics, 0, len(mS.Gauges)+len(mS.Counters))

	for name, value := range mS.Gauges {
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.MetricTypeGauge,
			Value: &value,
		})
	}

	for name, value := range mS.Counters {
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.MetricTypeCounter,
			Delta: &value,
		})
	}
	return metrics

}
