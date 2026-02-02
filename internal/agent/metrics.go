package agent

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"

	"github.com/F3dosik/metalert.git/pkg/models"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type Metrics struct {
	Gauges   map[string]models.Gauge
	Counters map[string]models.Counter
	mu       sync.RWMutex
}

func NewMetrics() *Metrics {
	return &Metrics{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}
}

var gettersRuntime = map[string]func(*runtime.MemStats) models.Gauge{
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

func (m *Metrics) UpdateGopsutilMetrics() {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return
	}

	cpuPercensts, err := cpu.Percent(0, true)
	if err != nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.Gauges["TotalMemory"] = models.Gauge(vm.Total)
	m.Gauges["FreeMemory"] = models.Gauge(vm.Free)

	for i, percent := range cpuPercensts {
		m.Gauges[fmt.Sprintf("CPUutilization%d", i+1)] = models.Gauge(percent)
	}
}

func (m *Metrics) UpdateRuntimeMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.mu.Lock()
	defer m.mu.Unlock()

	for name, getter := range gettersRuntime {
		m.Gauges[name] = getter(&memStats)
	}

	m.Gauges["RandomValue"] = models.Gauge(rand.Float64() * 100)
	m.Counters["PollCount"]++
}

type MetricsSnapshot struct {
	Gauges   map[string]models.Gauge
	Counters map[string]models.Counter
}

func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gauges := make(map[string]models.Gauge, len(m.Gauges))
	for k, v := range m.Gauges {
		gauges[k] = v
	}

	counters := make(map[string]models.Counter, len(m.Counters))
	for k, v := range m.Counters {
		counters[k] = v
	}

	return MetricsSnapshot{
		Gauges:   gauges,
		Counters: counters,
	}
}
