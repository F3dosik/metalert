package agent

import (
	"math/rand/v2"
	"runtime"
	"sync"

	"github.com/F3dosik/metalert.git/pkg/models"
)

type Metrics struct {
	Gauges   map[string]models.Gauge
	Counters map[string]models.Counter
	mu       sync.RWMutex
}

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

func (m *Metrics) Update() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Обновление метрик runtime
	for name, getter := range getters {
		m.Gauges[name] = getter(&memStats)
	}

	randomValue := rand.Float64() * 100
	m.Gauges["RandomValue"] = models.Gauge(randomValue)

	m.Counters["PollCount"]++
}
