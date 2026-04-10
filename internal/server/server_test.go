package server_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	cfg "github.com/F3dosik/metalert/internal/config/server"
	"github.com/F3dosik/metalert/internal/server"
)

func testLogger() *zap.SugaredLogger {
	l, _ := zap.NewDevelopment()
	return l.Sugar()
}

func minConfig() *cfg.ServerConfig {
	return &cfg.ServerConfig{
		Addr:     "localhost:0",
		AddrGRPC: "localhost:0",
		LogMode:  "development",
	}
}

func TestNewServer_MemStorage(t *testing.T) {
	s := server.NewServer(minConfig(), testLogger())
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestNewServer_TrustedSubnet(t *testing.T) {
	c := minConfig()
	c.TrustedSubnet = "192.168.0.0/16"
	s := server.NewServer(c, testLogger())
	if s == nil {
		t.Fatal("NewServer returned nil with trusted subnet")
	}
}

func TestNewServer_FileStorage(t *testing.T) {
	c := minConfig()
	c.FileStoragePath = t.TempDir() + "/metrics.json"
	s := server.NewServer(c, testLogger())
	if s == nil {
		t.Fatal("NewServer returned nil with file storage")
	}
}

func TestNewServer_WithStoreInterval(t *testing.T) {
	c := minConfig()
	c.FileStoragePath = t.TempDir() + "/metrics.json"
	c.StoreInterval = 300
	s := server.NewServer(c, testLogger())
	if s == nil {
		t.Fatal("NewServer returned nil with store interval")
	}
}

func TestAutoSave_CancelledContext(t *testing.T) {
	c := minConfig()
	c.StoreInterval = 1
	c.FileStoragePath = t.TempDir() + "/metrics.json"
	s := server.NewServer(c, testLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.AutoSave(ctx)
		close(done)
	}()

	select {
	case <-done:
		// AutoSave завершился после отмены контекста
	case <-time.After(2 * time.Second):
		t.Error("AutoSave did not stop after context cancellation")
	}
}

func TestAutoSave_MemStorageNoOp(t *testing.T) {
	// MemStorage не реализует Savable — AutoSave сразу возвращается.
	s := server.NewServer(minConfig(), testLogger())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // отменяем сразу

	done := make(chan struct{})
	go func() {
		s.AutoSave(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("AutoSave did not return for non-savable storage")
	}
}
