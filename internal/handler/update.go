package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/F3dosik/metalert/internal/audit"
	"github.com/F3dosik/metalert/internal/repository"
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
func UpdateHandler(storage repository.MetricsStorage, dispatcher *audit.AuditDispatcher, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		update(w, r, storage, dispatcher, logger)
	}
}

func update(
	w http.ResponseWriter, r *http.Request,
	storage repository.MetricsStorage, dispatcher *audit.AuditDispatcher,
	logger *zap.SugaredLogger,
) {
	var metName, metValue string

	metType := models.MetricType(chi.URLParam(r, "metType"))
	metName = chi.URLParam(r, "metName")
	metValue = chi.URLParam(r, "metValue")

	value, err := service.CheckAndParseValue(metType, metName, metValue)
	if err != nil {
		handleServiceError(w, logger, err)
		return
	}

	service.UpdateMetric(r.Context(), storage, metName, value)

	go dispatcher.Publish(audit.AuditEvent{
		Ts:        time.Now().Unix(),
		Metrics:   []string{metName},
		IPAddress: getIP(r),
	})

	message := fmt.Sprint("Метрика ", metName, " успешно обновлена\r\n")
	RespondTextOK(w, message)
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
// При asyncSave=true после обновления асинхронно вызывает Save() у хранилища,
// если оно реализует интерфейс repository.Savable.
//
// При успехе возвращает 200 OK. При ошибке — 400 Bad Request.
func UpdateJSONHandler(
	storage repository.MetricsStorage, dispatcher *audit.AuditDispatcher,
	logger *zap.SugaredLogger, asyncSave bool,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updateJSON(w, r, storage, dispatcher, logger, asyncSave)
	}
}

func updateJSON(
	w http.ResponseWriter, r *http.Request,
	storage repository.MetricsStorage, dispatcher *audit.AuditDispatcher,
	logger *zap.SugaredLogger, saveOnUpdate bool,
) {
	logger.Debug("decoding request")

	var metric models.Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		logger.Debug("cannot decode metric JSON body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := updateMetrics(r.Context(), storage, []models.Metric{metric}); err != nil {
		handleServiceError(w, logger, err)
		return
	}

	go dispatcher.Publish(audit.AuditEvent{
		Ts:        time.Now().Unix(),
		Metrics:   metricNames([]models.Metric{metric}),
		IPAddress: getIP(r),
	})

	if saveOnUpdate {
		if s, ok := storage.(repository.Savable); ok {
			go func() {
				if err := s.Save(); err != nil {
					logger.Warnw("error saving metrics", "error", err)
				}
			}()
		}
	}

	logger.Debug("sending HTTP 200 response")
	message := fmt.Sprint("Метрика ", metric.ID, " успешно обновлена\r\n")
	RespondTextOK(w, message)
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
// Для хранилища типа DBMetricsStorage все метрики обновляются в одной транзакции.
// При asyncSave=true после обновления асинхронно вызывает Save() у хранилища,
// если оно реализует интерфейс repository.Savable.
//
// При успехе возвращает 200 OK. При пустом массиве или ошибке — 400 Bad Request.
func UpdatesJSONHandler(
	storage repository.MetricsStorage, dispatcher *audit.AuditDispatcher,
	logger *zap.SugaredLogger, asyncSave bool,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updatesJSON(w, r, storage, dispatcher, logger, asyncSave)
	}
}

func updatesJSON(
	w http.ResponseWriter, r *http.Request,
	storage repository.MetricsStorage, dispatcher *audit.AuditDispatcher,
	logger *zap.SugaredLogger, saveOnUpdate bool,
) {
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
	if err := updateMetrics(r.Context(), storage, metrics); err != nil {
		handleServiceError(w, logger, err)
		return
	}

	go dispatcher.Publish(audit.AuditEvent{
		Ts:        time.Now().Unix(),
		Metrics:   metricNames(metrics),
		IPAddress: getIP(r),
	})

	if saveOnUpdate {
		if s, ok := storage.(repository.Savable); ok {
			go func() {
				if err := s.Save(); err != nil {
					logger.Warnw("error saving metrics", "error", err)
				}
			}()
		}
	}
	logger.Debug("sending HTTP 200 response")
	message := fmt.Sprint("Успешно обновлено ", len(metrics), " метрик\r\n")
	RespondTextOK(w, message)
}

func metricNames(metrics []models.Metric) []string {
	names := make([]string, len(metrics))
	for i, m := range metrics {
		names[i] = m.ID
	}
	return names
}

func getIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func updateMetrics(ctx context.Context, storage repository.MetricsStorage, metrics []models.Metric) error {
	for _, metric := range metrics {
		if err := validateMetric(metric); err != nil {
			return err
		}
	}

	switch s := storage.(type) {
	case *repository.DBMetricsStorage:
		if err := s.UpdateMetricTx(ctx, metrics); err != nil {
			return err
		}
	default:
		for _, metric := range metrics {
			if err := service.UpdateMetricFromStruct(ctx, s, metric); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateMetric(metric models.Metric) error {
	if err := metric.ValidateMeta(); err != nil {
		return err
	}
	if err := metric.ValidateValue(); err != nil {
		return err
	}
	return nil
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
