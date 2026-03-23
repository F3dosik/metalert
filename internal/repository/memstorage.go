package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/F3dosik/metalert/pkg/models"
)

// MemMetricsStorage — потокобезопасное in-memory хранилище метрик.
//
// Хранит gauge- и counter-метрики в обычных map-ах, защищённых RWMutex.
// Данные не переживают перезапуск процесса.
// Для персистентного хранения используйте [FileMetricsStorage].
//
// Пример:
//
//	storage := repository.NewMemMetricsStorage()
//	storage.SetGauge(ctx, "cpu", 72.5)
//	val, _ := storage.GetGauge(ctx, "cpu")
type MemMetricsStorage struct {
	// Gauges содержит все gauge-метрики. Публичное поле позволяет
	// сериализовать хранилище напрямую (например, в FileMetricsStorage).
	Gauges map[string]models.Gauge

	// Counters содержит все counter-метрики.
	Counters map[string]models.Counter

	mutex sync.RWMutex
}

// NewMemMetricsStorage создаёт новое пустое in-memory хранилище метрик.
func NewMemMetricsStorage() *MemMetricsStorage {
	return &MemMetricsStorage{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}
}

// SetGauge устанавливает значение gauge-метрики. Существующее значение перезаписывается.
func (f *MemMetricsStorage) SetGauge(ctx context.Context, metName string, value models.Gauge) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.Gauges[metName] = value
	return nil
}

// GetGauge возвращает текущее значение gauge-метрики по имени.
// Возвращает ошибку, если метрика не найдена.
func (f *MemMetricsStorage) GetGauge(ctx context.Context, metName string) (models.Gauge, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	v, ok := f.Gauges[metName]
	if !ok {
		return 0, fmt.Errorf("метрика %s отсутствует", metName)
	}

	return v, nil
}

// AddCounter прибавляет value к текущему значению counter-метрики.
// Если метрика не существует, создаётся с переданным значением.
func (f *MemMetricsStorage) AddCounter(ctx context.Context, metName string, value models.Counter) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.Counters[metName] += value
	return nil
}

// SetCounter устанавливает значение counter-метрики, заменяя текущее.
// В отличие от AddCounter, не прибавляет, а перезаписывает.
func (f *MemMetricsStorage) SetCounter(ctx context.Context, metName string, value models.Counter) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.Counters[metName] = value
	return nil
}

// GetCounter возвращает текущее значение counter-метрики по имени.
// Возвращает ошибку, если метрика не найдена.
func (f *MemMetricsStorage) GetCounter(ctx context.Context, metName string) (models.Counter, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	v, ok := f.Counters[metName]
	if !ok {
		return 0, fmt.Errorf("метрика %s отсутствует", metName)
	}

	return v, nil
}

// GetAllMetrics возвращает снимок всех метрик из хранилища в произвольном порядке.
func (f *MemMetricsStorage) GetAllMetrics(ctx context.Context) ([]models.Metric, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	metrics := make([]models.Metric, 0, len(f.Gauges)+len(f.Counters))

	for name, value := range f.Gauges {
		v := value
		metrics = append(metrics, models.Metric{
			ID:    name,
			MType: models.TypeGauge,
			Value: &v,
		})
	}

	for name, value := range f.Counters {
		d := value
		metrics = append(metrics, models.Metric{
			ID:    name,
			MType: models.TypeCounter,
			Delta: &d,
		})
	}

	return metrics, nil
}
