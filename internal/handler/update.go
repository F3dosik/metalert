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

func UpdateHandler(storage *repository.MemMetricsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		update(w, r, storage)
	}
}

func update(w http.ResponseWriter, r *http.Request, storage *repository.MemMetricsStorage) {
	// if !isPlainText(r) {
	// 	http.Error(rw, errInvalidContentType.Error(), http.StatusBadRequest)
	// 	return
	// }

	var metName, metValue string

	metType := models.MetricType(chi.URLParam(r, "metType"))
	metName = chi.URLParam(r, "metName")
	metValue = chi.URLParam(r, "metValue")

	value, err := service.CheckAndParseValue(metType, metName, metValue)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	service.UpdateMetric(storage, metName, value)

	message := fmt.Sprintf("Метрика %s успешно обновлена\r\n", metName)
	RespondTextOK(w, message)

}

func UpdateJSONHandler(storage *repository.MemMetricsStorage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updateJSON(w, r, storage, logger)
	}
}

func updateJSON(w http.ResponseWriter, r *http.Request, storage *repository.MemMetricsStorage, logger *zap.SugaredLogger) {
	logger.Debug("decoding request")

	var metric models.Metric
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metric); err != nil {
		logger.Debug("cannot decode metric JSON body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := metric.ValidateMeta(); err != nil {
		logger.Debug("invalid metric meta", zap.Error(err))
		handleServiceError(w, err)
		return
	}

	if err := metric.ValidateValue(); err != nil {
		logger.Debug("invalid metric value", zap.Error(err))
		handleServiceError(w, err)
		return
	}

	service.UpdateMetricFromStruct(storage, metric)

	logger.Debug("sending HTTP 200 response")
	message := fmt.Sprintf("Метрика %s успешно обновлена\r\n", metric.ID)
	RespondTextOK(w, message)

}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, models.ErrInvalidType):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, models.ErrNoName):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, models.ErrInvalidValue):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, models.ErrInvalidDelta):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
