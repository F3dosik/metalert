package service

import (
	"testing"

	"github.com/F3dosik/metalert/pkg/models"
)

func TestCheckAndParseValue(t *testing.T) {
	tests := []struct {
		name      string
		metType   models.MetricType
		metName   string
		metValue  string
		wantErr   error
		wantGauge *models.Gauge
		wantCount *models.Counter
	}{
		{
			name:      "valid gauge",
			metType:   models.TypeGauge,
			metName:   "cpu",
			metValue:  "72.5",
			wantGauge: gaugePtr(72.5),
		},
		{
			name:      "valid gauge integer",
			metType:   models.TypeGauge,
			metName:   "mem",
			metValue:  "100",
			wantGauge: gaugePtr(100),
		},
		{
			name:      "valid counter",
			metType:   models.TypeCounter,
			metName:   "hits",
			metValue:  "42",
			wantCount: counterPtr(42),
		},
		{
			name:     "invalid type",
			metType:  "unknown",
			metName:  "x",
			metValue: "1",
			wantErr:  models.ErrInvalidType,
		},
		{
			name:     "empty name",
			metType:  models.TypeGauge,
			metName:  "",
			metValue: "1.0",
			wantErr:  models.ErrNoName,
		},
		{
			name:     "invalid gauge value",
			metType:  models.TypeGauge,
			metName:  "cpu",
			metValue: "not-a-number",
			wantErr:  models.ErrInvalidValue,
		},
		{
			name:     "invalid counter value",
			metType:  models.TypeCounter,
			metName:  "hits",
			metValue: "3.14",
			wantErr:  models.ErrInvalidDelta,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CheckAndParseValue(tt.metType, tt.metName, tt.metValue)
			if err != tt.wantErr {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantGauge != nil {
				g, ok := result.(models.Gauge)
				if !ok {
					t.Fatalf("expected Gauge result, got %T", result)
				}
				if g != *tt.wantGauge {
					t.Errorf("gauge = %v, want %v", g, *tt.wantGauge)
				}
			}
			if tt.wantCount != nil {
				c, ok := result.(models.Counter)
				if !ok {
					t.Fatalf("expected Counter result, got %T", result)
				}
				if c != *tt.wantCount {
					t.Errorf("counter = %v, want %v", c, *tt.wantCount)
				}
			}
		})
	}
}

func TestValidateMetricType(t *testing.T) {
	if err := ValidateMetricType(models.TypeGauge); err != nil {
		t.Errorf("gauge: unexpected error %v", err)
	}
	if err := ValidateMetricType(models.TypeCounter); err != nil {
		t.Errorf("counter: unexpected error %v", err)
	}
	if err := ValidateMetricType("bad"); err != models.ErrInvalidType {
		t.Errorf("bad type: want ErrInvalidType, got %v", err)
	}
}

func TestValidateMetric(t *testing.T) {
	v := models.Gauge(1.5)
	d := models.Counter(10)

	tests := []struct {
		name    string
		metric  models.Metric
		wantErr error
	}{
		{"valid gauge", models.Metric{ID: "cpu", MType: models.TypeGauge, Value: &v}, nil},
		{"valid counter", models.Metric{ID: "req", MType: models.TypeCounter, Delta: &d}, nil},
		{"missing id", models.Metric{ID: "", MType: models.TypeGauge, Value: &v}, models.ErrNoName},
		{"invalid type", models.Metric{ID: "x", MType: "bad"}, models.ErrInvalidType},
		{"gauge nil value", models.Metric{ID: "x", MType: models.TypeGauge}, models.ErrInvalidValue},
		{"counter nil delta", models.Metric{ID: "x", MType: models.TypeCounter}, models.ErrInvalidDelta},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetric(tt.metric)
			if err != tt.wantErr {
				t.Errorf("ValidateMetric() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func gaugePtr(v float64) *models.Gauge   { g := models.Gauge(v); return &g }
func counterPtr(v int64) *models.Counter { c := models.Counter(v); return &c }
