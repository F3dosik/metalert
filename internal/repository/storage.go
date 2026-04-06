// Package repository содержит реализации хранилища метрик.
//
// Поддерживаются три варианта хранилища:
//   - [MemMetricsStorage] — in-memory, данные теряются при перезапуске
//   - [FileMetricsStorage] — персистентное файловое хранилище поверх MemMetricsStorage
//   - [DBMetricsStorage] — PostgreSQL с поддержкой транзакций и retry
//
// Все реализации удовлетворяют интерфейсу [MetricsStorage].
package repository

import (
	"context"

	"github.com/F3dosik/metalert/pkg/models"
)

// MetricsStorage — основной интерфейс хранилища метрик.
//
// Все реализации (MemMetricsStorage, FileMetricsStorage, DBMetricsStorage)
// обязаны реализовывать этот интерфейс.
type MetricsStorage interface {
	// SetGauge устанавливает значение gauge-метрики с заданным именем.
	// Если метрика уже существует, значение перезаписывается.
	SetGauge(ctx context.Context, name string, value models.Gauge) error

	// GetGauge возвращает текущее значение gauge-метрики по имени.
	// Возвращает ошибку, если метрика не найдена.
	GetGauge(ctx context.Context, name string) (models.Gauge, error)

	// AddCounter прибавляет value к текущему значению counter-метрики.
	// Если метрика не существует, создаётся с переданным значением.
	AddCounter(ctx context.Context, name string, value models.Counter) error

	// GetCounter возвращает текущее значение counter-метрики по имени.
	// Возвращает ошибку, если метрика не найдена.
	GetCounter(ctx context.Context, name string) (models.Counter, error)

	// GetAllMetrics возвращает все метрики из хранилища.
	GetAllMetrics(ctx context.Context) ([]models.Metric, error)

	// UpdateMany атомарно обновляет набор метрик.
	// Реализации вправе выполнять обновление в одной транзакции (DB) или последовательно (memory/file).
	UpdateMany(ctx context.Context, metrics []models.Metric) error
}

// Savable — дополнительный интерфейс для хранилищ, поддерживающих сброс на диск.
//
// Реализуется FileMetricsStorage. Используется хендлерами для асинхронного
// сохранения после каждого обновления метрики (если включён режим asyncSave).
type Savable interface {
	// Save сохраняет текущее состояние хранилища в постоянное хранилище.
	Save() error
}
