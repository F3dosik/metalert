// Package repository содержит реализацию хранилища метрик.
// Внутреннее хранилище MemStorage позволяет сохранять
// значения типа Gauge и Counter, а также обновлять их.
package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/F3dosik/metalert.git/pkg/models"
)

type MetricsStorage interface {
	SetGauge(name string, value models.Gauge)
	GetGauge(name string) (models.Gauge, error)

	SetCounter(name string, value models.Counter)
	AddCounter(name string, value models.Counter)
	GetCounter(name string) (models.Counter, error)

	GetAllMetrics() []models.Metric
}

type MemMetricsStorage struct {
	Gauges   map[string]models.Gauge
	Counters map[string]models.Counter
	mutex    sync.RWMutex
	fileName string
}

func NewMemMetricsStorage(fileName string) *MemMetricsStorage {
	return &MemMetricsStorage{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
		mutex:    sync.RWMutex{},
		fileName: fileName,
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

func (mS *MemMetricsStorage) AddCounter(metName string, value models.Counter) {
	mS.mutex.Lock()
	defer mS.mutex.Unlock()
	mS.Counters[metName] += value

}

func (mS *MemMetricsStorage) SetCounter(metName string, value models.Counter) {
	mS.mutex.Lock()
	defer mS.mutex.Unlock()

	mS.Counters[metName] = value
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

func (mS *MemMetricsStorage) Save() error {
	metrics := mS.GetAllMetrics()

	data, err := json.Marshal(&metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics: %w", err)
	}

	tmpFile := mS.fileName + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0666); err != nil {
		return fmt.Errorf("write tmp file: %w", err)
	}

	return os.Rename(tmpFile, mS.fileName)
}

func (mS *MemMetricsStorage) Load() error {
	data, err := os.ReadFile(mS.fileName)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var metrics []models.Metric
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("unmarshal metrics: %w", err)
	}

	for _, metric := range metrics {
		switch metric.MType {
		case models.TypeGauge:
			mS.SetGauge(metric.ID, *metric.Value)
		case models.TypeCounter:
			mS.SetCounter(metric.ID, *metric.Delta)
		}
	}

	return nil
}
