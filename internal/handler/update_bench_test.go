package handler

import (
	"context"
	"testing"

	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/pkg/models"
)

func gaugeMetric(name string, v float64) models.Metric {
	val := models.Gauge(v)
	return models.Metric{ID: name, MType: models.TypeGauge, Value: &val}
}

func counterMetric(name string, d int64) models.Metric {
	delta := models.Counter(d)
	return models.Metric{ID: name, MType: models.TypeCounter, Delta: &delta}
}

func makeBatch(n int) []models.Metric {
	metrics := make([]models.Metric, n)
	for i := range metrics {
		if i%2 == 0 {
			metrics[i] = gaugeMetric("Alloc", float64(i)*1.5)
		} else {
			metrics[i] = counterMetric("Requests", int64(i))
		}
	}
	return metrics
}

// --- validateMetric ---

func BenchmarkValidateMetric_Gauge(b *testing.B) {
	m := gaugeMetric("Alloc", 123.45)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		validateMetric(m)
	}
}

func BenchmarkValidateMetric_Counter(b *testing.B) {
	m := counterMetric("Requests", 42)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		validateMetric(m)
	}
}

func BenchmarkValidateMetric_InvalidType(b *testing.B) {
	m := models.Metric{ID: "X", MType: "unknown"}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		validateMetric(m)
	}
}

// --- updateMetrics ---

func BenchmarkUpdateMetrics_Single(b *testing.B) {
	s := repository.NewMemMetricsStorage()
	ctx := context.Background()
	metrics := makeBatch(1)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		updateMetrics(ctx, s, metrics)
	}
}

func BenchmarkUpdateMetrics_Batch10(b *testing.B) {
	s := repository.NewMemMetricsStorage()
	ctx := context.Background()
	metrics := makeBatch(10)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		updateMetrics(ctx, s, metrics)
	}
}

func BenchmarkUpdateMetrics_Batch100(b *testing.B) {
	s := repository.NewMemMetricsStorage()
	ctx := context.Background()
	metrics := makeBatch(100)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		updateMetrics(ctx, s, metrics)
	}
}
