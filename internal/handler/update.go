package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/F3dosik/metalert/internal/service"
	"github.com/F3dosik/metalert/pkg/models"
)

// UpdateHandler возвращает HTTP-хендлер для обновления метрики через URL-параметры.
//
// Маршрут: POST /update/{metType}/{metName}/{metValue}
//
// Параметры пути:
//   - metType  — тип метрики ("gauge" или "counter")
//   - metName  — имя метрики
//   - metValue — новое значение метрики
//
// При успехе возвращает 200 OK с текстовым подтверждением.
// При ошибке валидации — 400 Bad Request или 404 Not Found.
func UpdateHandler(svc service.MetricsService, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		update(w, r, svc, logger)
	}
}

func update(w http.ResponseWriter, r *http.Request, svc service.MetricsService, logger *zap.SugaredLogger) {
	metType := models.MetricType(chi.URLParam(r, "metType"))
	metName := chi.URLParam(r, "metName")
	metValue := chi.URLParam(r, "metValue")

	if err := svc.Update(r.Context(), metType, metName, metValue, getIP(r)); err != nil {
		handleServiceError(w, logger, err)
		return
	}

	RespondTextOK(w, fmt.Sprint("Метрика ", metName, " успешно обновлена\r\n"))
}

// UpdateJSONHandler возвращает HTTP-хендлер для обновления одной метрики через JSON.
//
// Маршрут: POST /update/
//
// Тело запроса — JSON-объект типа models.Metric:
//
//	{"id": "cpu", "type": "gauge", "value": 72.5}
//	{"id": "requests", "type": "counter", "delta": 1}
//
// При успехе возвращает 200 OK. При ошибке — 400 Bad Request.
func UpdateJSONHandler(svc service.MetricsService, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updateJSON(w, r, svc, logger)
	}
}

func updateJSON(w http.ResponseWriter, r *http.Request, svc service.MetricsService, logger *zap.SugaredLogger) {
	logger.Debug("decoding request")

	var metric models.Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		logger.Debug("cannot decode metric JSON body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := svc.UpdateMany(r.Context(), []models.Metric{metric}, getIP(r)); err != nil {
		handleServiceError(w, logger, err)
		return
	}

	logger.Debug("sending HTTP 200 response")
	RespondTextOK(w, fmt.Sprint("Метрика ", metric.ID, " успешно обновлена\r\n"))
}

// UpdatesJSONHandler возвращает HTTP-хендлер для пакетного обновления метрик через JSON.
//
// Маршрут: POST /updates/
//
// Тело запроса — JSON-массив объектов типа models.Metric:
//
//	[
//	  {"id": "cpu", "type": "gauge", "value": 72.5},
//	  {"id": "requests", "type": "counter", "delta": 10}
//	]
//
// При успехе возвращает 200 OK. При пустом массиве или ошибке — 400 Bad Request.
func UpdatesJSONHandler(svc service.MetricsService, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updatesJSON(w, r, svc, logger)
	}
}

func updatesJSON(w http.ResponseWriter, r *http.Request, svc service.MetricsService, logger *zap.SugaredLogger) {
	logger.Debug("decoding request")

	var metrics []models.Metric
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		logger.Warnw("cannot decode metrics JSON body", "err", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if len(metrics) == 0 {
		http.Error(w, "empty metrics array", http.StatusBadRequest)
		return
	}

	if err := svc.UpdateMany(r.Context(), metrics, getIP(r)); err != nil {
		handleServiceError(w, logger, err)
		return
	}

	logger.Debug("sending HTTP 200 response")
	RespondTextOK(w, fmt.Sprint("Успешно обновлено ", len(metrics), " метрик\r\n"))
}

func getIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// handleServiceError переводит ошибки сервисного слоя в соответствующие HTTP-ответы:
//   - ErrInvalidType, ErrInvalidValue, ErrInvalidDelta → 400 Bad Request
//   - ErrNoName → 404 Not Found
//   - остальные → 500 Internal Server Error
func handleServiceError(w http.ResponseWriter, logger *zap.SugaredLogger, err error) {
	switch {
	case errors.Is(err, models.ErrInvalidType),
		errors.Is(err, models.ErrInvalidValue),
		errors.Is(err, models.ErrInvalidDelta):
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Debugw("bad metric input", "error", err)

	case errors.Is(err, models.ErrNoName):
		http.Error(w, err.Error(), http.StatusNotFound)
		logger.Debugw("metric name not provided", "error", err)

	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		logger.Errorw("internal server error", "error", err)
	}
}
