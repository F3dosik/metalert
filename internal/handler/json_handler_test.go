package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/F3dosik/metalert/internal/audit"
	"github.com/F3dosik/metalert/internal/repository"
	"github.com/F3dosik/metalert/pkg/models"
)

func testLog() *zap.SugaredLogger {
	l, _ := zap.NewDevelopment()
	return l.Sugar()
}

func setup(t *testing.T) (*repository.MemMetricsStorage, *audit.AuditDispatcher, *zap.SugaredLogger) {
	t.Helper()
	return repository.NewMemMetricsStorage(), &audit.AuditDispatcher{}, testLog()
}

// ── UpdateJSONHandler ────────────────────────────────────────────────────────

func TestUpdateJSONHandler_Gauge(t *testing.T) {
	storage, dispatcher, sugar := setup(t)
	_ = sugar

	log := testLog()
	h := UpdateJSONHandler(storage, dispatcher, log, false)

	v := models.Gauge(42.0)
	body, _ := json.Marshal(models.Metric{ID: "cpu", MType: models.TypeGauge, Value: &v})

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateJSONHandler_Counter(t *testing.T) {
	storage, dispatcher, _ := setup(t)
	log := testLog()
	h := UpdateJSONHandler(storage, dispatcher, log, false)

	d := models.Counter(7)
	body, _ := json.Marshal(models.Metric{ID: "req", MType: models.TypeCounter, Delta: &d})

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
}

func TestUpdateJSONHandler_InvalidJSON(t *testing.T) {
	storage, dispatcher, _ := setup(t)
	log := testLog()
	h := UpdateJSONHandler(storage, dispatcher, log, false)

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString("bad json"))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestUpdateJSONHandler_InvalidMetric(t *testing.T) {
	storage, dispatcher, _ := setup(t)
	log := testLog()
	h := UpdateJSONHandler(storage, dispatcher, log, false)

	// Gauge without value → validation error
	body, _ := json.Marshal(models.Metric{ID: "cpu", MType: models.TypeGauge})
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

// ── UpdatesJSONHandler ───────────────────────────────────────────────────────

func TestUpdatesJSONHandler_Batch(t *testing.T) {
	storage, dispatcher, _ := setup(t)
	log := testLog()
	h := UpdatesJSONHandler(storage, dispatcher, log, false)

	v := models.Gauge(1.1)
	d := models.Counter(5)
	metrics := []models.Metric{
		{ID: "cpu", MType: models.TypeGauge, Value: &v},
		{ID: "req", MType: models.TypeCounter, Delta: &d},
	}
	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdatesJSONHandler_EmptyArray(t *testing.T) {
	storage, dispatcher, _ := setup(t)
	log := testLog()
	h := UpdatesJSONHandler(storage, dispatcher, log, false)

	body, _ := json.Marshal([]models.Metric{})
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400 for empty array, got %d", rr.Code)
	}
}

func TestUpdatesJSONHandler_InvalidJSON(t *testing.T) {
	storage, dispatcher, _ := setup(t)
	log := testLog()
	h := UpdatesJSONHandler(storage, dispatcher, log, false)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString("bad"))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

// ── ValueJSONHandler ─────────────────────────────────────────────────────────

func TestValueJSONHandler_Gauge(t *testing.T) {
	storage, _, _ := setup(t)
	log := testLog()
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()
	_ = storage.SetGauge(ctx, "cpu", 72.5)

	h := ValueJSONHandler(storage, log)
	body, _ := json.Marshal(models.Metric{ID: "cpu", MType: models.TypeGauge})
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
	var resp models.Metric
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Value == nil || *resp.Value != 72.5 {
		t.Errorf("value = %v, want 72.5", resp.Value)
	}
}

func TestValueJSONHandler_Counter(t *testing.T) {
	storage, _, _ := setup(t)
	log := testLog()
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()
	_ = storage.AddCounter(ctx, "hits", 42)

	h := ValueJSONHandler(storage, log)
	body, _ := json.Marshal(models.Metric{ID: "hits", MType: models.TypeCounter})
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
	var resp models.Metric
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Delta == nil || *resp.Delta != 42 {
		t.Errorf("delta = %v, want 42", resp.Delta)
	}
}

func TestValueJSONHandler_NotFound(t *testing.T) {
	storage, _, _ := setup(t)
	log := testLog()
	h := ValueJSONHandler(storage, log)

	body, _ := json.Marshal(models.Metric{ID: "missing", MType: models.TypeGauge})
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rr.Code)
	}
}

func TestValueJSONHandler_InvalidJSON(t *testing.T) {
	storage, _, _ := setup(t)
	log := testLog()
	h := ValueJSONHandler(storage, log)

	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBufferString("bad"))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestValueJSONHandler_InvalidMeta(t *testing.T) {
	storage, _, _ := setup(t)
	log := testLog()
	h := ValueJSONHandler(storage, log)

	body, _ := json.Marshal(models.Metric{ID: "", MType: models.TypeGauge})
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("want 404 for ErrNoName, got %d", rr.Code)
	}
}

// ── ValueHandler (URL) ───────────────────────────────────────────────────────

func TestValueHandler_GaugeFound(t *testing.T) {
	storage, _, _ := setup(t)
	log := testLog()
	_ = log
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()
	_ = storage.SetGauge(ctx, "temp", 36.6)

	r := chi.NewRouter()
	r.Get("/value/{metType}/{metName}", ValueHandler(storage))

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/temp", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
}

func TestValueHandler_NotFound(t *testing.T) {
	storage, _, _ := setup(t)
	r := chi.NewRouter()
	r.Get("/value/{metType}/{metName}", ValueHandler(storage))

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/missing", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rr.Code)
	}
}

func TestValueHandler_InvalidType(t *testing.T) {
	storage, _, _ := setup(t)
	r := chi.NewRouter()
	r.Get("/value/{metType}/{metName}", ValueHandler(storage))

	req := httptest.NewRequest(http.MethodGet, "/value/badtype/x", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}
