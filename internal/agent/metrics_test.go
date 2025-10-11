package agent

import (
	"testing"

	"github.com/F3dosik/metalert.git/pkg/models"
)

func TestUpdateMetrics(t *testing.T) {
	metrics := &Metrics{
		Gauges:   make(map[string]models.Gauge),
		Counters: make(map[string]models.Counter),
	}

	metrics.Update()
	// Проверяем наличие всех метрик после обновления.
	t.Run("Gauges", func(t *testing.T) {
		for name := range getters {
			if _, ok := metrics.Gauges[name]; !ok {
				t.Errorf("Отсутствует метрика: %s", name)
			}
		}
	})

	t.Run("RandomValue", func(t *testing.T) {
		if _, ok := metrics.Gauges["RandomValue"]; !ok {
			t.Error("Отсутствует метрика: RandomValue")
		}
	})

	t.Run("PollCount increment", func(t *testing.T) {
		metrics = &Metrics{
			Gauges:   make(map[string]models.Gauge),
			Counters: make(map[string]models.Counter),
		}

		metrics.Update()
		if metrics.Counters["PollCount"] != 1 {
			t.Errorf("PollCount = %d; хотели 1", metrics.Counters["PollCount"])
		}

		metrics.Update()
		if metrics.Counters["PollCount"] != 2 {
			t.Errorf("PollCount = %d; хотели 2", metrics.Counters["PollCount"])
		}
	})

}
