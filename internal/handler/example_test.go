package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/F3dosik/metalert.git/internal/audit"
	"github.com/F3dosik/metalert.git/internal/handler"
	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/pkg/models"
)

// noopDispatcher — заглушка аудит-диспетчера, не делающая ничего.
func newNoopDispatcher() *audit.AuditDispatcher {
	return audit.NewAuditDispatcher(zap.NewNop().Sugar())
}

// newRouter собирает chi-роутер с хендлерами сервера метрик.
func newRouter(storage *repository.MemMetricsStorage) http.Handler {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	dispatcher := newNoopDispatcher()

	r := chi.NewRouter()
	r.Post("/update/{metType}/{metName}/{metValue}", handler.UpdateHandler(storage, dispatcher, sugar))
	r.Post("/update/", handler.UpdateJSONHandler(storage, dispatcher, sugar, false))
	r.Post("/updates/", handler.UpdatesJSONHandler(storage, dispatcher, sugar, false))
	r.Get("/value/{metType}/{metName}", handler.ValueHandler(storage))
	r.Post("/value/", handler.ValueJSONHandler(storage, sugar))

	return r
}

// Example демонстрирует полный цикл работы с сервером метрик:
// обновление gauge и counter через URL и JSON, пакетное обновление,
// а также получение значений через URL и JSON.
func Example() {
	storage := repository.NewMemMetricsStorage()
	srv := httptest.NewServer(newRouter(storage))
	defer srv.Close()

	// 1. Обновление gauge-метрики через URL-параметры
	// POST /update/gauge/cpu_usage/72.5
	resp, _ := http.Post(srv.URL+"/update/gauge/cpu_usage/72.5", "", nil)
	fmt.Println("UpdateHandler gauge:", resp.StatusCode)
	resp.Body.Close()

	// 2. Обновление counter-метрики через URL-параметры
	// POST /update/counter/requests_total/1
	resp, _ = http.Post(srv.URL+"/update/counter/requests_total/1", "", nil)
	fmt.Println("UpdateHandler counter:", resp.StatusCode)
	resp.Body.Close()

	// 3. Обновление одной метрики через JSON
	// POST /update/
	gauge := models.NewMetricGauge("memory_usage", 512.0)
	body, _ := json.Marshal(gauge)
	resp, _ = http.Post(srv.URL+"/update/", "application/json", bytes.NewReader(body))
	fmt.Println("UpdateJSONHandler:", resp.StatusCode)
	resp.Body.Close()

	// 4. Пакетное обновление метрик через JSON
	// POST /updates/
	metrics := []models.Metric{
		*models.NewMetricGauge("disk_usage", 88.3),
		*models.NewMetricCounter("errors_total", 5),
	}
	body, _ = json.Marshal(metrics)
	resp, _ = http.Post(srv.URL+"/updates/", "application/json", bytes.NewReader(body))
	fmt.Println("UpdatesJSONHandler:", resp.StatusCode)
	resp.Body.Close()

	// 5. Получение gauge-метрики через URL
	// GET /value/gauge/cpu_usage
	resp, _ = http.Get(srv.URL + "/value/gauge/cpu_usage")
	val, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println("ValueHandler gauge:", string(val))

	// 6. Получение counter-метрики через URL
	// GET /value/counter/requests_total
	resp, _ = http.Get(srv.URL + "/value/counter/requests_total")
	val, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println("ValueHandler counter:", string(val))

	// 7. Получение метрики через JSON
	// POST /value/
	req := models.Metric{ID: "memory_usage", MType: models.TypeGauge}
	body, _ = json.Marshal(req)
	resp, _ = http.Post(srv.URL+"/value/", "application/json", bytes.NewReader(body))
	var result models.Metric
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	fmt.Printf("ValueJSONHandler: id=%s value=%.1f\n", result.ID, *result.Value)

	// Output:
	// UpdateHandler gauge: 200
	// UpdateHandler counter: 200
	// UpdateJSONHandler: 200
	// UpdatesJSONHandler: 200
	// ValueHandler gauge: 72.5
	// ValueHandler counter: 1
	// ValueJSONHandler: id=memory_usage value=512.0
}
