package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func testLogger() *zap.SugaredLogger {
	l, _ := zap.NewDevelopment()
	return l.Sugar()
}

// compressBody сжимает данные для тела запроса.
func compressBody(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(data)
	_ = zw.Close()
	return buf.Bytes()
}

func TestWithCompression_PlainRequest(t *testing.T) {
	// Клиент не поддерживает и не отправляет gzip — запрос проходит без изменений.
	h := WithCompression(testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("plain body"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
	if rr.Body.String() != "plain body" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "plain body")
	}
}

func TestWithCompression_AcceptGzip(t *testing.T) {
	// Клиент объявляет Accept-Encoding: gzip — ответ должен быть сжат.
	h := WithCompression(testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Error("want Content-Encoding: gzip")
	}

	// Проверяем, что тело можно расжать.
	zr, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	defer zr.Close()
	got, _ := io.ReadAll(zr)
	if string(got) != "hello" {
		t.Errorf("decompressed = %q, want %q", got, "hello")
	}
}

func TestWithCompression_SendsGzip(t *testing.T) {
	// Клиент отправляет тело сжатым — middleware должен расжать перед хендлером.
	h := WithCompression(testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	}))

	compressed := compressBody(t, []byte("compressed payload"))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(compressed))
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
	if rr.Body.String() != "compressed payload" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "compressed payload")
	}
}

func TestWithCompression_InvalidGzipBody(t *testing.T) {
	// Клиент утверждает gzip, но данные не валидны — должен вернуть 500.
	h := WithCompression(testLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("should not reach"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not gzip"))
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("want 500 for invalid gzip, got %d", rr.Code)
	}
}

func TestCompressWriter_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	cw := newCompressWriter(rr)
	cw.WriteHeader(http.StatusCreated)
	_ = cw.Close()

	if rr.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rr.Code)
	}
}

func TestCompressWriter_Header(t *testing.T) {
	rr := httptest.NewRecorder()
	cw := newCompressWriter(rr)
	cw.Header().Set("X-Test", "value")
	_ = cw.Close()

	if rr.Header().Get("X-Test") != "value" {
		t.Error("header not forwarded")
	}
}
