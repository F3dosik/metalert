package handler

import (
	"context"
	"testing"

	"github.com/F3dosik/metalert/internal/repository"
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

func newBenchService(b *testing.B) service.MetricsService {
	b.Helper()
	return service.NewMetricsService(repository.NewMemMetricsStorage(), nil, false, testLog())
}

// --- UpdateMany ---

func BenchmarkUpdateMetrics_Single(b *testing.B) {
	svc := newBenchService(b)
	ctx := context.Background()
	metrics := makeBatch(b, 1)
	b.ReportAllocs()

	for b.Loop() {
		svc.UpdateMany(ctx, metrics, "")
	}
}

func BenchmarkUpdateMetrics_Batch10(b *testing.B) {
	svc := newBenchService(b)
	ctx := context.Background()
	metrics := makeBatch(b, 10)
	b.ReportAllocs()

	for b.Loop() {
		svc.UpdateMany(ctx, metrics, "")
	}
}

func BenchmarkUpdateMetrics_Batch100(b *testing.B) {
	svc := newBenchService(b)
	ctx := context.Background()
	metrics := makeBatch(b, 100)
	b.ReportAllocs()

	for b.Loop() {
		svc.UpdateMany(ctx, metrics, "")
	}
}
