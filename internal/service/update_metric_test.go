package service

import (
	"context"
	"testing"

	"github.com/F3dosik/metalert/internal/repository"
	"github.com/F3dosik/metalert/pkg/models"
)

func TestUpdateMetric_Gauge(t *testing.T) {
	storage := repository.NewMemMetricsStorage()
	ctx := context.Background()

	if err := UpdateMetric(ctx, storage, "cpu", models.Gauge(55.5)); err != nil {
		t.Fatalf("UpdateMetric gauge: %v", err)
	}

	got, err := storage.GetGauge(ctx, "cpu")
	if err != nil {
		t.Fatalf("GetGauge: %v", err)
	}
	if got != 55.5 {
		t.Errorf("gauge = %v, want 55.5", got)
	}
}

func TestUpdateMetric_Counter(t *testing.T) {
	storage := repository.NewMemMetricsStorage()
	ctx := context.Background()

	if err := UpdateMetric(ctx, storage, "hits", models.Counter(10)); err != nil {
		t.Fatalf("UpdateMetric counter: %v", err)
	}
	if err := UpdateMetric(ctx, storage, "hits", models.Counter(5)); err != nil {
		t.Fatalf("UpdateMetric counter second: %v", err)
	}

	got, err := storage.GetCounter(ctx, "hits")
	if err != nil {
		t.Fatalf("GetCounter: %v", err)
	}
	if got != 15 {
		t.Errorf("counter = %v, want 15", got)
	}
}

func TestUpdateMetric_UnknownType(t *testing.T) {
	storage := repository.NewMemMetricsStorage()
	// Неизвестный тип — UpdateMetric просто ничего не делает, ошибку не возвращает.
	if err := UpdateMetric(context.Background(), storage, "x", "unexpected-type"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateMetricFromStruct_Gauge(t *testing.T) {
	storage := repository.NewMemMetricsStorage()
	ctx := context.Background()
	v := models.Gauge(3.14)

	if err := UpdateMetricFromStruct(ctx, storage, models.Metric{
		ID: "pi", MType: models.TypeGauge, Value: &v,
	}); err != nil {
		t.Fatalf("UpdateMetricFromStruct gauge: %v", err)
	}

	got, _ := storage.GetGauge(ctx, "pi")
	if got != 3.14 {
		t.Errorf("got %v, want 3.14", got)
	}
}

func TestUpdateMetricFromStruct_Counter(t *testing.T) {
	storage := repository.NewMemMetricsStorage()
	ctx := context.Background()
	d := models.Counter(7)

	if err := UpdateMetricFromStruct(ctx, storage, models.Metric{
		ID: "reqs", MType: models.TypeCounter, Delta: &d,
	}); err != nil {
		t.Fatalf("UpdateMetricFromStruct counter: %v", err)
	}

	got, _ := storage.GetCounter(ctx, "reqs")
	if got != 7 {
		t.Errorf("got %v, want 7", got)
	}
}

func TestUpdateMetrics_ValidBatch(t *testing.T) {
	storage := repository.NewMemMetricsStorage()
	ctx := context.Background()

	v := models.Gauge(1.1)
	d := models.Counter(5)
	metrics := []models.Metric{
		{ID: "cpu", MType: models.TypeGauge, Value: &v},
		{ID: "req", MType: models.TypeCounter, Delta: &d},
	}

	if err := UpdateMetrics(ctx, storage, metrics); err != nil {
		t.Fatalf("UpdateMetrics: %v", err)
	}

	if g, _ := storage.GetGauge(ctx, "cpu"); g != 1.1 {
		t.Errorf("gauge = %v, want 1.1", g)
	}
	if c, _ := storage.GetCounter(ctx, "req"); c != 5 {
		t.Errorf("counter = %v, want 5", c)
	}
}

func TestUpdateMetrics_InvalidMetric(t *testing.T) {
	storage := repository.NewMemMetricsStorage()
	// Метрика без имени — должна вернуть ошибку.
	metrics := []models.Metric{
		{ID: "", MType: models.TypeGauge},
	}
	if err := UpdateMetrics(context.Background(), storage, metrics); err == nil {
		t.Error("expected error for invalid metric")
	}
}

func TestUpdateMetrics_EmptyBatch(t *testing.T) {
	storage := repository.NewMemMetricsStorage()
	if err := UpdateMetrics(context.Background(), storage, nil); err != nil {
		t.Errorf("unexpected error for empty batch: %v", err)
	}
}
