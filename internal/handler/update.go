// Package handler содержит HTTP-хэндлеры сервера метрик.
// Здесь реализованы функции обработки POST-запросов,
// проверки URL, типов и значений метрик и возврата ответов клиенту.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	"net/http"

	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/internal/service"
	"github.com/F3dosik/metalert.git/pkg/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func UpdateHandler(storage repository.MetricsStorage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		update(w, r, storage, logger)
	}
}

func update(w http.ResponseWriter, r *http.Request, storage repository.MetricsStorage, logger *zap.SugaredLogger) {

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

	message := fmt.Sprint("Метрика ", metName, " успешно обновлена\r\n")
	RespondTextOK(w, message)

}

func UpdateJSONHandler(storage repository.MetricsStorage, logger *zap.SugaredLogger, asyncSave bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updateJSON(w, r, storage, logger, asyncSave)
	}
}

func updateJSON(w http.ResponseWriter, r *http.Request, storage repository.MetricsStorage, logger *zap.SugaredLogger, saveOnUpdate bool) {
	logger.Debug("decoding request")

	var metric models.Metric
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metric); err != nil {
		logger.Debug("cannot decode metric JSON body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := metric.ValidateMeta(); err != nil {
		handleServiceError(w, logger, err)
		return
	}

	if err := metric.ValidateValue(); err != nil {
		handleServiceError(w, logger, err)
		return
	}

	service.UpdateMetricFromStruct(r.Context(), storage, metric)

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

func handleServiceError(w http.ResponseWriter, logger *zap.SugaredLogger, err error) {
	switch {
	case errors.Is(err, models.ErrInvalidType):
		logger.Debugw("invalid metric type", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)

	case errors.Is(err, models.ErrNoName):
		logger.Debugw("metric name not provided", "error", err)
		http.Error(w, err.Error(), http.StatusNotFound)

	case errors.Is(err, models.ErrInvalidValue),
		errors.Is(err, models.ErrInvalidDelta):
		logger.Debugw("invalid metric value", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)

	default:
		logger.Errorw("internal server error", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
