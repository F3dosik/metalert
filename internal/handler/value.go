package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/F3dosik/metalert/internal/service"
	"github.com/F3dosik/metalert/pkg/models"
)

// ValueHandler возвращает HTTP-хендлер для получения значения метрики через URL-параметры.
//
// Маршрут: GET /value/{metType}/{metName}
//
// При успехе возвращает 200 OK с текстовым значением метрики.
// При неизвестном типе — 400 Bad Request.
// При отсутствии метрики — 404 Not Found.
func ValueHandler(svc service.MetricsService) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		value(rw, r, svc)
	}
}

func value(rw http.ResponseWriter, r *http.Request, svc service.MetricsService) {
	metType := models.MetricType(chi.URLParam(r, "metType"))
	metName := chi.URLParam(r, "metName")
	ctx := r.Context()

	if err := service.ValidateMetricType(metType); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	var message string
	switch metType {
	case models.TypeGauge:
		val, err := svc.GetGauge(ctx, metName)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		message = strconv.FormatFloat(float64(val), 'f', -1, 64)

	case models.TypeCounter:
		val, err := svc.GetCounter(ctx, metName)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		message = strconv.Itoa(int(val))
	}
	RespondTextOK(rw, message)
}

// ValueJSONHandler возвращает HTTP-хендлер для получения значения метрики через JSON.
//
// Маршрут: POST /value/
//
// Тело запроса — JSON-объект с полями id и type:
//
//	{"id": "cpu", "type": "gauge"}
//	{"id": "requests", "type": "counter"}
//
// Возвращает тот же объект с заполненным полем Value (для gauge) или Delta (для counter).
// При невалидных данных — 400 Bad Request.
// При отсутствии метрики — 404 Not Found.
func ValueJSONHandler(svc service.MetricsService, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		valueJSON(w, r, svc, logger)
	}
}

func valueJSON(w http.ResponseWriter, r *http.Request, svc service.MetricsService, logger *zap.SugaredLogger) {
	logger.Debug("decoding request")

	var metric models.Metric
	ctx := r.Context()
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		logger.Debug("cannot decode metric JSON body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := metric.ValidateMeta(); err != nil {
		logger.Debug("invalid metric", zap.Error(err))
		handleServiceError(w, logger, err)
		return
	}

	response := metric
	switch metric.MType {
	case models.TypeGauge:
		val, err := svc.GetGauge(ctx, metric.ID)
		if err != nil {
			logger.Debug("cannot GetGauge", zap.Error(err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		response.Value = &val

	case models.TypeCounter:
		val, err := svc.GetCounter(ctx, metric.ID)
		if err != nil {
			logger.Debug("cannot GetCounter", zap.Error(err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		response.Delta = &val
	}

	RespondJSONOK(w, response)
}
