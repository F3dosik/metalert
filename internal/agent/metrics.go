package agent

import (
	"math/rand/v2"
	"runtime"
	"sync"

	"github.com/F3dosik/metalert/pkg/models"
)

// Metrics хранит текущий снимок метрик агента.
//
// Поле mu защищает Gauges и Counters от гонки данных:
// [Metrics.Update] захватывает полный Lock, [Sender.SendMetrics] — RLock.
type Metrics struct {
	// Gauges содержит gauge-метрики: значения runtime.MemStats и RandomValue.
	Gauges map[string]models.Gauge

	// Counters содержит counter-метрики. На данный момент единственная
	// метрика — PollCount, инкрементируемая при каждом вызове Update.
	Counters map[string]models.Counter

	mu sync.RWMutex
}

// getters — таблица функций-извлекателей gauge-метрик из runtime.MemStats.
// Каждый ключ соответствует имени метрики, отправляемой на сервер.
var getters = map[string]func(*runtime.MemStats) models.Gauge{
	"Alloc":         func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.Alloc) },
	"BuckHashSys":   func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.BuckHashSys) },
	"Frees":         func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.Frees) },
	"GCCPUFraction": func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.GCCPUFraction) },
	"GCSys":         func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.GCSys) },
	"HeapAlloc":     func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapAlloc) },
	"HeapIdle":      func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapIdle) },
	"HeapInuse":     func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapInuse) },
	"HeapObjects":   func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapObjects) },
	"HeapReleased":  func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapReleased) },
	"HeapSys":       func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.HeapSys) },
	"LastGC":        func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.LastGC) },
	"Lookups":       func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.Lookups) },
	"MCacheInuse":   func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.MCacheInuse) },
	"MCacheSys":     func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.MCacheSys) },
	"MSpanSys":      func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.MSpanSys) },
	"MSpanInuse":    func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.MSpanInuse) },
	"Mallocs":       func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.Mallocs) },
	"NextGC":        func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.NextGC) },
	"NumForcedGC":   func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.NumForcedGC) },
	"NumGC":         func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.NumGC) },
	"OtherSys":      func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.OtherSys) },
	"PauseTotalNs":  func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.PauseTotalNs) },
	"StackInuse":    func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.StackInuse) },
	"StackSys":      func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.StackSys) },
	"Sys":           func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.Sys) },
	"TotalAlloc":    func(m *runtime.MemStats) models.Gauge { return models.Gauge(m.TotalAlloc) },
}

// Update собирает текущий снимок runtime.MemStats и обновляет все gauge-метрики.
//
// Дополнительно устанавливает:
//   - RandomValue — случайное число в диапазоне [0, 100)
//   - PollCount — счётчик, инкрементируемый при каждом вызове
//
// Потокобезопасен: захватывает полный mutex на время записи.
func (m *Metrics) Update() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.mu.Lock()
	defer m.mu.Unlock()

	for name, getter := range getters {
		m.Gauges[name] = getter(&memStats)
	}

	m.Gauges["RandomValue"] = models.Gauge(rand.Float64() * 100)
	m.Counters["PollCount"]++
}
