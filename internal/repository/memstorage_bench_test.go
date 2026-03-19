package repository_test

import (
	"context"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/pkg/models"
)

func BenchmarkMemStorage_SetGauge(b *testing.B) {
	s := repository.NewMemMetricsStorage()
	ctx := context.Background()
	b.ResetTimer()

	for b.Loop() {
		s.SetGauge(ctx, "Alloc", models.Gauge(rand.Float64()))
	}
}

func BenchmarkMemStorage_AddCounter(b *testing.B) {
	s := repository.NewMemMetricsStorage()
	ctx := context.Background()
	b.ResetTimer()

	for b.Loop() {
		s.AddCounter(ctx, "Requests", models.Counter(rand.Int64()))
	}
}

func BenchmarkMemStorage_GetGauge_Hit(b *testing.B) {
	s := repository.NewMemMetricsStorage()
	ctx := context.Background()
	s.SetGauge(ctx, "Alloc", 42.0)
	b.ResetTimer()

	for b.Loop() {
		s.GetGauge(ctx, "Alloc")
	}
}

func BenchmarkMemStorage_GetGauge_Miss(b *testing.B) {
	s := repository.NewMemMetricsStorage()
	ctx := context.Background()
	b.ResetTimer()

	for b.Loop() {
		s.GetGauge(ctx, "nonexistent")
	}
}

func BenchmarkMemStorage_GetAllMetrics(b *testing.B) {
	s := repository.NewMemMetricsStorage()
	ctx := context.Background()

	// Наполняем реалистичными данными
	for i := 0; i < 30; i++ {
		s.SetGauge(ctx, "gauge_"+strconv.Itoa(i), models.Gauge(float64(i)))
	}
	for i := 0; i < 10; i++ {
		s.AddCounter(ctx, "counter_"+strconv.Itoa(i), models.Counter(i))
	}
	b.ResetTimer()

	for b.Loop() {
		s.GetAllMetrics(ctx)
	}
}
