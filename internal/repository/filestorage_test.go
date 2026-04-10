package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/F3dosik/metalert/pkg/models"
)

func TestNewFileMetricsStorage_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.json")

	fs, err := NewFileMetricsStorage(path, false)
	if err != nil {
		t.Fatalf("NewFileMetricsStorage: %v", err)
	}
	defer fs.Close()

	if fs.Gauges == nil || fs.Counters == nil {
		t.Error("maps should be initialized")
	}
}

func TestFileMetricsStorage_SaveAndRestore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.json")
	ctx := context.Background()

	// Создаём хранилище, записываем метрики, сохраняем и закрываем.
	fs, err := NewFileMetricsStorage(path, false)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_ = fs.SetGauge(ctx, "cpu", 77.7)
	_ = fs.AddCounter(ctx, "req", 5)

	if err := fs.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Открываем снова с restore=true — данные должны восстановиться.
	fs2, err := NewFileMetricsStorage(path, true)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	defer fs2.Close()

	g, err := fs2.GetGauge(ctx, "cpu")
	if err != nil || g != 77.7 {
		t.Errorf("gauge = %v, err = %v, want 77.7", g, err)
	}
	c, err := fs2.GetCounter(ctx, "req")
	if err != nil || c != 5 {
		t.Errorf("counter = %v, err = %v, want 5", c, err)
	}
}

func TestFileMetricsStorage_Save(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.json")
	ctx := context.Background()

	fs, err := NewFileMetricsStorage(path, false)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer fs.Close()

	_ = fs.SetGauge(ctx, "mem", 512.0)

	if err := fs.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// После Save временный файл должен содержать JSON с метриками.
	tmpPath := path + ".tmp"
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("read tmp: %v", err)
	}

	var metrics []models.Metric
	if err := json.Unmarshal(data, &metrics); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(metrics) != 1 || metrics[0].ID != "mem" {
		t.Errorf("unexpected metrics: %+v", metrics)
	}
}

func TestNewFileMetricsStorage_RestoreNoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")

	// restore=true, но файл не существует — не должно быть фатальной ошибки.
	fs, err := NewFileMetricsStorage(path, true)
	if err != nil {
		t.Fatalf("NewFileMetricsStorage: %v", err)
	}
	defer fs.Close()
}

func TestNewFileMetricsStorage_NestedDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "metrics.json")

	fs, err := NewFileMetricsStorage(path, false)
	if err != nil {
		t.Fatalf("nested dir: %v", err)
	}
	defer fs.Close()
}
