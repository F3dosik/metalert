// Package repository содержит реализацию хранилища метрик.
// Внутреннее хранилище MemStorage позволяет сохранять
// значения типа Gauge и Counter, а также обновлять их.
package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/F3dosik/metalert.git/pkg/models"
)

type MemMetricsStorage struct {
	Gauges   map[string]models.Gauge
	Counters map[string]models.Counter
	mutex    sync.RWMutex
	fileName string
	tmpFile  *os.File
	ErrCh    chan error
}

func NewMemMetricsStorage(fileName string, restore bool) (*MemMetricsStorage, error) {
	absPath, err := filepath.Abs(fileName)
	if err != nil {
		absPath = fileName
	}

	dir := filepath.Dir(absPath)
	if err = os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("ошибка создания каталога %s: %w", dir, err)
	}

	tmpFile, err := os.OpenFile(absPath+".tmp", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	storage := &MemMetricsStorage{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
		mutex:    sync.RWMutex{},
		fileName: absPath,
		tmpFile:  tmpFile,
		ErrCh:    make(chan error, 10),
	}

	if restore {
		if err := storage.load(); err != nil {
			storage.ErrCh <- fmt.Errorf("не удалось восстановить метрики: %w", err)
		}
	}

	go storage.periodicSave(1 * time.Minute)

	return storage, nil

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

	if err := mS.tmpFile.Truncate(0); err != nil {
		return err
	}
	if _, err := mS.tmpFile.Seek(0, 0); err != nil {
		return err
	}

	if _, err := mS.tmpFile.Write(data); err != nil {
		return fmt.Errorf("write tmp file: %w", err)
	}

	return nil
}

func (mS *MemMetricsStorage) load() error {
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

func (mS *MemMetricsStorage) Close() error {
	if err := mS.Save(); err != nil {
		return fmt.Errorf("save before close: %w", err)
	}

	if err := mS.tmpFile.Sync(); err != nil {
		return fmt.Errorf("sync tmp file: %w", err)
	}

	if err := mS.tmpFile.Close(); err != nil {
		return fmt.Errorf("close tmp file: %w", err)
	}

	if err := os.Rename(mS.fileName+".tmp", mS.fileName); err != nil {
		return fmt.Errorf("rename tmp file: %w", err)
	}

	return nil
}

func (mS *MemMetricsStorage) periodicSave(interval time.Duration) {
	var err error
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err = mS.Close(); err != nil {
			mS.sendErr(err)
		}

		file, err := os.OpenFile(mS.fileName+".tmp", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			mS.sendErr(err)
		} else {
			mS.tmpFile = file
		}
	}
}

func (mS *MemMetricsStorage) sendErr(err error) {
	select {
	case mS.ErrCh <- err:
	default:
	}
}
