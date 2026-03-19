package handler

import (
	"context"
	"testing"

	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/pkg/models"
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

func makeBatch(b *testing.B, n int) []models.Metric {
	b.Helper()
	metrics := make([]models.Metric, n)
	for i := range metrics {
		if i%2 == 0 {
			metrics[i] = gaugeMetric(b, "Alloc", float64(i)*1.5)
		} else {
			metrics[i] = counterMetric(b, "Requests", int64(i))
		}
	}
	return metrics
}

// --- validateMetric ---

func BenchmarkValidateMetric_Gauge(b *testing.B) {
	m := gaugeMetric(b, "Alloc", 123.45)
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		validateMetric(m)
	}
}

func BenchmarkValidateMetric_Counter(b *testing.B) {
	m := counterMetric(b, "Requests", 42)
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		validateMetric(m)
	}
}

func BenchmarkValidateMetric_InvalidType(b *testing.B) {
	m := models.Metric{ID: "X", MType: "unknown"}
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		validateMetric(m)
	}
}

// --- updateMetrics ---

func BenchmarkUpdateMetrics_Single(b *testing.B) {
	s := repository.NewMemMetricsStorage()
	ctx := context.Background()
	metrics := makeBatch(b, 1)
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		updateMetrics(ctx, s, metrics)
	}
}

func BenchmarkUpdateMetrics_Batch10(b *testing.B) {
	s := repository.NewMemMetricsStorage()
	ctx := context.Background()
	metrics := makeBatch(b, 10)
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		updateMetrics(ctx, s, metrics)
	}
}

func BenchmarkUpdateMetrics_Batch100(b *testing.B) {
	s := repository.NewMemMetricsStorage()
	ctx := context.Background()
	metrics := makeBatch(b, 100)
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		updateMetrics(ctx, s, metrics)
	}
}
