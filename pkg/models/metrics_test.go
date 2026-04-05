package models

import (
	"testing"
)

func TestNewMetricGauge(t *testing.T) {
	m := NewMetricGauge("cpu", 72.5)
	if m.ID != "cpu" {
		t.Errorf("ID = %q, want %q", m.ID, "cpu")
	}
	if m.MType != TypeGauge {
		t.Errorf("MType = %q, want %q", m.MType, TypeGauge)
	}
	if m.Value == nil || *m.Value != 72.5 {
		t.Errorf("Value = %v, want 72.5", m.Value)
	}
	if m.Delta != nil {
		t.Error("Delta should be nil for gauge")
	}
}

func TestNewMetricCounter(t *testing.T) {
	m := NewMetricCounter("hits", 42)
	if m.ID != "hits" {
		t.Errorf("ID = %q, want %q", m.ID, "hits")
	}
	if m.MType != TypeCounter {
		t.Errorf("MType = %q, want %q", m.MType, TypeCounter)
	}
	if m.Delta == nil || *m.Delta != 42 {
		t.Errorf("Delta = %v, want 42", m.Delta)
	}
	if m.Value != nil {
		t.Error("Value should be nil for counter")
	}
}

func TestIsValidMetricType(t *testing.T) {
	tests := []struct {
		input MetricType
		want  bool
	}{
		{TypeGauge, true},
		{TypeCounter, true},
		{"unknown", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsValidMetricType(tt.input); got != tt.want {
			t.Errorf("IsValidMetricType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestValidateMeta(t *testing.T) {
	tests := []struct {
		name    string
		metric  Metric
		wantErr error
	}{
		{"valid gauge", Metric{ID: "cpu", MType: TypeGauge}, nil},
		{"valid counter", Metric{ID: "req", MType: TypeCounter}, nil},
		{"empty id", Metric{ID: "", MType: TypeGauge}, ErrNoName},
		{"invalid type", Metric{ID: "x", MType: "bad"}, ErrInvalidType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.ValidateMeta()
			if err != tt.wantErr {
				t.Errorf("ValidateMeta() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue(t *testing.T) {
	v := Gauge(1.0)
	d := Counter(1)

	tests := []struct {
		name    string
		metric  Metric
		wantErr error
	}{
		{"gauge with value", Metric{MType: TypeGauge, Value: &v}, nil},
		{"gauge nil value", Metric{MType: TypeGauge, Value: nil}, ErrInvalidValue},
		{"counter with delta", Metric{MType: TypeCounter, Delta: &d}, nil},
		{"counter nil delta", Metric{MType: TypeCounter, Delta: nil}, ErrInvalidDelta},
		{"invalid type", Metric{MType: "bad"}, ErrInvalidType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.ValidateValue()
			if err != tt.wantErr {
				t.Errorf("ValidateValue() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
