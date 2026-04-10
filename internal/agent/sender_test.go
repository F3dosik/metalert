package agent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	grpcserver "github.com/F3dosik/metalert/internal/grpc"
	pb "github.com/F3dosik/metalert/internal/proto"
	"github.com/F3dosik/metalert/internal/repository"
	"github.com/F3dosik/metalert/pkg/models"
)

// newTestSender создаёт Sender, направленный на тестовый httptest-сервер.
func newTestSender(serverURL string) *Sender {
	return NewSender(serverURL, "", nil)
}

func newMetrics() *Metrics {
	m := &Metrics{
		Gauges:   map[string]models.Gauge{"cpu": 50.5, "mem": 80.0},
		Counters: map[string]models.Counter{"hits": 10, "errors": 2},
	}
	return m
}

// ── resolveLocalIP ────────────────────────────────────────────────────────────

func TestResolveLocalIP_ValidURL(t *testing.T) {
	ip := resolveLocalIP("http://localhost:8080")
	// Функция не должна паниковать; результат зависит от ОС, просто проверяем не пусто.
	_ = ip
}

func TestResolveLocalIP_EmptyURL(t *testing.T) {
	ip := resolveLocalIP("")
	if ip != "" {
		t.Errorf("expected empty for empty url, got %q", ip)
	}
}

func TestResolveLocalIP_InvalidURL(t *testing.T) {
	ip := resolveLocalIP("://bad url")
	_ = ip // не должен паниковать
}

// ── prepareURL ────────────────────────────────────────────────────────────────

func TestPrepareURL_AddsScheme(t *testing.T) {
	s := newTestSender("localhost:8080")
	got := s.prepareURL("/updates/")
	if got == "" {
		t.Error("prepareURL returned empty")
	}
}

func TestPrepareURL_WithScheme(t *testing.T) {
	s := newTestSender("http://localhost:8080")
	got := s.prepareURL("/updates/")
	want := "http://localhost:8080/updates/"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// ── encryptIfNeeded ───────────────────────────────────────────────────────────

func TestEncryptIfNeeded_NoKey(t *testing.T) {
	s := newTestSender("http://localhost:8080")
	data := []byte("payload")
	out, encrypted, err := s.encryptIfNeeded(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if encrypted {
		t.Error("expected not encrypted when no key")
	}
	if string(out) != "payload" {
		t.Errorf("data changed unexpectedly: %q", out)
	}
}

// ── sendMetricsBatch / sendMetricJSON / sendMetricsIndividually ───────────────

func TestSendMetricsBatch_Success(t *testing.T) {
	var received []models.Metric
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/updates/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := newTestSender(srv.URL)
	m := newMetrics()
	m.mu.RLock()
	defer m.mu.RUnlock()
	s.sendMetricsBatch(m, false)

	if len(received) != 4 {
		t.Errorf("expected 4 metrics, got %d", len(received))
	}
}

func TestSendMetricsBatch_WithCompression(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Error("expected gzip Content-Encoding")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := newTestSender(srv.URL)
	m := newMetrics()
	m.mu.RLock()
	defer m.mu.RUnlock()
	s.sendMetricsBatch(m, true)
}

func TestSendMetricsBatch_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := newTestSender(srv.URL)
	m := newMetrics()
	m.mu.RLock()
	defer m.mu.RUnlock()
	s.sendMetricsBatch(m, false) // should log but not panic
}

func TestSendMetricsIndividually_Success(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := newTestSender(srv.URL)
	m := newMetrics()
	m.mu.RLock()
	defer m.mu.RUnlock()
	s.sendMetricsIndividually(m)

	if calls != 4 {
		t.Errorf("expected 4 calls (one per metric), got %d", calls)
	}
}

func TestSendMetricURL_NotOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	s := newTestSender(srv.URL)
	err := s.sendMetricURL(models.TypeGauge, "cpu", "50")
	if err == nil {
		t.Error("expected error for 400 response")
	}
}

// ── SendMetrics (public) ──────────────────────────────────────────────────────

func TestSendMetrics_JSON(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := newTestSender(srv.URL)
	m := newMetrics()
	s.SendMetrics(m, "JSON", false)

	if !called {
		t.Error("server was not called")
	}
}

func TestSendMetrics_URL(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := newTestSender(srv.URL)
	m := newMetrics()
	s.SendMetrics(m, "URL", false)

	if calls == 0 {
		t.Error("server was not called for URL mode")
	}
}

func TestSendMetrics_UnknownType(t *testing.T) {
	s := newTestSender("http://localhost:1")
	m := newMetrics()
	s.SendMetrics(m, "UNKNOWN", false) // should just log, no panic
}

func TestSendMetrics_GRPC_NoClient(t *testing.T) {
	s := newTestSender("http://localhost:1")
	m := newMetrics()
	s.SendMetrics(m, "GRPC", false) // grpcClient is nil, should log and return
}

// ── NewSender ─────────────────────────────────────────────────────────────────

func TestNewSender_Basic(t *testing.T) {
	s := NewSender("http://localhost:8080", "", nil)
	if s.ServerURL != "http://localhost:8080" {
		t.Errorf("ServerURL = %q", s.ServerURL)
	}
	if s.Client == nil {
		t.Error("Client is nil")
	}
	if s.grpcClient != nil {
		t.Error("grpcClient should be nil without grpc endpoint")
	}
}

func TestNewSender_WithGRPC(t *testing.T) {
	// grpc.NewClient с невалидным адресом — должен создать клиент (соединение ленивое)
	s := NewSender("http://localhost:8080", "localhost:9999", nil)
	if s.grpcClient == nil {
		t.Error("grpcClient should be non-nil when grpc endpoint given")
	}
}

// ── sendMetricsGRPC через bufconn ─────────────────────────────────────────────

func TestSendMetricsGRPC_Success(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	storage := repository.NewMemMetricsStorage()
	srv := grpc.NewServer()
	pb.RegisterMetricsServer(srv, grpcserver.NewMetricsServer(storage))
	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}
	defer conn.Close()

	s := newTestSender("http://localhost:8080")
	s.grpcClient = pb.NewMetricsClient(conn)

	m := newMetrics()
	m.mu.RLock()
	defer m.mu.RUnlock()
	s.sendMetricsGRPC(m)
}

func TestSendMetrics_GRPC_WithClient(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	storage := repository.NewMemMetricsStorage()
	srv := grpc.NewServer()
	pb.RegisterMetricsServer(srv, grpcserver.NewMetricsServer(storage))
	go srv.Serve(lis)
	defer srv.Stop()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}
	defer conn.Close()

	s := newTestSender("http://localhost:8080")
	s.grpcClient = pb.NewMetricsClient(conn)

	m := newMetrics()
	s.SendMetrics(m, "GRPC", false)
}

// ── encryptIfNeeded с ключом ──────────────────────────────────────────────────

func TestEncryptIfNeeded_WithKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	s := newTestSender("http://localhost:8080")
	s.CryptoKey = &privateKey.PublicKey

	data := []byte(`[{"id":"cpu","type":"gauge","value":50.0}]`)
	out, encrypted, err := s.encryptIfNeeded(data)
	if err != nil {
		t.Fatalf("encryptIfNeeded: %v", err)
	}
	if !encrypted {
		t.Error("expected encrypted=true with key set")
	}
	if len(out) == 0 {
		t.Error("encrypted output is empty")
	}
}
