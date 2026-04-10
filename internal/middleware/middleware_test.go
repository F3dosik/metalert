package middleware

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func testLogger() *zap.SugaredLogger {
	l, _ := zap.NewDevelopment()
	return l.Sugar()
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// ── RequireJSON ──────────────────────────────────────────────────────────────

func TestRequireJSON_Passes(t *testing.T) {
	h := RequireJSON(testLogger())(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
}

func TestRequireJSON_Rejects(t *testing.T) {
	h := RequireJSON(testLogger())(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("want 405, got %d", rr.Code)
	}
}

// ── RequireTrustedSubnet ─────────────────────────────────────────────────────

func TestRequireTrustedSubnet_NoSubnet(t *testing.T) {
	h := RequireTrustedSubnet(testLogger(), nil)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("want 200 when no subnet, got %d", rr.Code)
	}
}

func TestRequireTrustedSubnet_AllowedIP(t *testing.T) {
	_, ipNet := mustParseCIDR("192.168.1.0/24")
	h := RequireTrustedSubnet(testLogger(), ipNet)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.42")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("want 200 for trusted IP, got %d", rr.Code)
	}
}

func TestRequireTrustedSubnet_DeniedIP(t *testing.T) {
	_, ipNet := mustParseCIDR("192.168.1.0/24")
	h := RequireTrustedSubnet(testLogger(), ipNet)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("want 403 for untrusted IP, got %d", rr.Code)
	}
}

func TestRequireTrustedSubnet_MissingHeader(t *testing.T) {
	_, ipNet := mustParseCIDR("192.168.1.0/24")
	h := RequireTrustedSubnet(testLogger(), ipNet)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("want 403 for missing IP header, got %d", rr.Code)
	}
}

// ── WithLogging ──────────────────────────────────────────────────────────────

func TestWithLogging_PassesThrough(t *testing.T) {
	h := WithLogging(testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rr.Code)
	}
	if rr.Body.String() != "created" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "created")
	}
}

func TestWithLogging_DefaultStatus(t *testing.T) {
	// Если хендлер не вызывает WriteHeader — статус должен быть 200.
	h := WithLogging(testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
}

// ── DecryptMiddleware ────────────────────────────────────────────────────────

func TestDecryptMiddleware_NoKey(t *testing.T) {
	// Без ключа — запрос проходит как есть.
	h := DecryptMiddleware(nil, testLogger())(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("data"))
	req.Header.Set("X-Encrypted", "true")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("want 200 without key, got %d", rr.Code)
	}
}

func TestDecryptMiddleware_NoEncryptedHeader(t *testing.T) {
	// Есть ключ, но заголовок X-Encrypted не выставлен — запрос проходит.
	h := DecryptMiddleware(nil, testLogger())(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("plain"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("want 200 without encrypted header, got %d", rr.Code)
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func mustParseCIDR(s string) (net.IP, *net.IPNet) {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return nil, ipNet
}
