package repository

import (
	"context"
	"testing"

	"github.com/F3dosik/metalert/pkg/models"
)

func TestNewMemMetricsStorage(t *testing.T) {
	s := NewMemMetricsStorage()
	if s.Gauges == nil {
		t.Error("Gauges map is nil")
	}
	if s.Counters == nil {
		t.Error("Counters map is nil")
	}
}

func TestGetGauge_Found(t *testing.T) {
	s := NewMemMetricsStorage()
	ctx := context.Background()

	_ = s.SetGauge(ctx, "temp", 36.6)

	got, err := s.GetGauge(ctx, "temp")
	if err != nil {
		t.Fatalf("GetGauge: %v", err)
	}
	if got != 36.6 {
		t.Errorf("got %v, want 36.6", got)
	}
}

func TestGetGauge_NotFound(t *testing.T) {
	s := NewMemMetricsStorage()
	_, err := s.GetGauge(context.Background(), "missing")
	if err == nil {
		t.Error("expected error for missing gauge")
	}
}

func TestGetCounter_Found(t *testing.T) {
	s := NewMemMetricsStorage()
	ctx := context.Background()

	_ = s.AddCounter(ctx, "hits", 10)
	_ = s.AddCounter(ctx, "hits", 5)

	got, err := s.GetCounter(ctx, "hits")
	if err != nil {
		t.Fatalf("GetCounter: %v", err)
	}
	if got != 15 {
		t.Errorf("got %v, want 15", got)
	}
}

func TestGetCounter_NotFound(t *testing.T) {
	s := NewMemMetricsStorage()
	_, err := s.GetCounter(context.Background(), "missing")
	if err == nil {
		t.Error("expected error for missing counter")
	}
}

func TestSetCounter(t *testing.T) {
	s := NewMemMetricsStorage()
	ctx := context.Background()

	_ = s.AddCounter(ctx, "req", 100)
	_ = s.SetCounter(ctx, "req", 1) // перезаписывает, не складывает

	got, _ := s.GetCounter(ctx, "req")
	if got != 1 {
		t.Errorf("SetCounter: got %v, want 1", got)
	}
}

func TestGetAllMetrics(t *testing.T) {
	s := NewMemMetricsStorage()
	ctx := context.Background()

	_ = s.SetGauge(ctx, "cpu", 50.0)
	_ = s.SetGauge(ctx, "mem", 80.0)
	_ = s.AddCounter(ctx, "reqs", 42)

	all, err := s.GetAllMetrics(ctx)
	if err != nil {
		t.Fatalf("GetAllMetrics: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("len = %d, want 3", len(all))
	}

	byID := make(map[string]models.Metric)
	for _, m := range all {
		byID[m.ID] = m
	}

	if m, ok := byID["cpu"]; !ok || m.MType != models.TypeGauge || *m.Value != 50.0 {
		t.Errorf("cpu metric wrong: %+v", byID["cpu"])
	}
	if m, ok := byID["reqs"]; !ok || m.MType != models.TypeCounter || *m.Delta != 42 {
		t.Errorf("reqs metric wrong: %+v", byID["reqs"])
	}
}

func TestGetAllMetrics_Empty(t *testing.T) {
	s := NewMemMetricsStorage()
	all, err := s.GetAllMetrics(context.Background())
	if err != nil {
		t.Fatalf("GetAllMetrics: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected empty, got %d", len(all))
	}
}
