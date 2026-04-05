package service_test

import (
	"testing"

	"github.com/F3dosik/metalert/internal/service"
	"github.com/F3dosik/metalert/pkg/models"
)

func gaugeMetric(b *testing.B, name string, v float64) models.Metric {
	b.Helper()
	val := models.Gauge(v)
	return models.Metric{ID: name, MType: models.TypeGauge, Value: &val}
}

func counterMetric(b *testing.B, name string, d int64) models.Metric {
	b.Helper()
	delta := models.Counter(d)
	return models.Metric{ID: name, MType: models.TypeCounter, Delta: &delta}
}
func BenchmarkCheckAndParseValue_Gauge(b *testing.B) {
	for b.Loop() {
		service.CheckAndParseValue(models.TypeGauge, "Alloc", "123.45")
	}
}

func BenchmarkCheckAndParseValue_Counter(b *testing.B) {
	for b.Loop() {
		service.CheckAndParseValue(models.TypeCounter, "Requests", "42")
	}
}

func BenchmarkCheckAndParseValue_InvalidType(b *testing.B) {
	for b.Loop() {
		service.CheckAndParseValue("unknown", "Alloc", "123.45")
	}
}

func BenchmarkValidateMetric_Gauge(b *testing.B) {
	m := gaugeMetric(b, "Alloc", 123.45)
	b.ReportAllocs()

	for b.Loop() {
		service.ValidateMetric(m)
	}
}

func BenchmarkValidateMetric_Counter(b *testing.B) {
	m := counterMetric(b, "Requests", 42)
	b.ReportAllocs()

	for b.Loop() {
		service.ValidateMetric(m)
	}
}

func BenchmarkValidateMetric_InvalidType(b *testing.B) {
	m := models.Metric{ID: "X", MType: "unknown"}
	b.ReportAllocs()

	for b.Loop() {
		service.ValidateMetric(m)
	}
}
