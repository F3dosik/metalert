package repository

import (
	"testing"

	models "github.com/F3dosik/metalert.git/internal/model"
)

func TestUpdateGauge(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricValue models.Gauge
		initial     map[string]models.Gauge
		want        models.Gauge
	}{
		{
			name:        "перезаписывает существующее значение",
			metricName:  "temp",
			metricValue: models.Gauge(40),
			initial:     map[string]models.Gauge{"temp": 38},
			want:        models.Gauge(40),
		},
		{
			name:        "оздает новое значение",
			metricName:  "temp",
			metricValue: models.Gauge(40),
			initial:     map[string]models.Gauge{},
			want:        models.Gauge(40),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := MemStorage{
				Gauges: map[string]models.Gauge{},
			}
			for k, v := range tt.initial {
				storage.Gauges[k] = v
			}

			storage.UpdateGauge(tt.metricName, tt.metricValue)

			if got := storage.Gauges[tt.metricName]; got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateCounter(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricValue models.Counter
		initial     map[string]models.Counter
		want        models.Counter
	}{
		{
			name:        "добавляет значение к предыдущему",
			metricName:  "count",
			metricValue: models.Counter(40),
			initial:     map[string]models.Counter{"count": 38},
			want:        models.Counter(78),
		},
		{
			name:        "инициализирует новое значение",
			metricName:  "new",
			metricValue: models.Counter(10),
			initial:     map[string]models.Counter{},
			want:        models.Counter(10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := MemStorage{
				Counters: map[string]models.Counter{},
			}
			for k, v := range tt.initial {
				storage.Counters[k] = v
			}

			storage.UpdateCounter(tt.metricName, tt.metricValue)

			if got := storage.Counters[tt.metricName]; got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
