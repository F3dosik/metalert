package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/internal/service"
	"github.com/F3dosik/metalert.git/pkg/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func ValueHandler(storage *repository.MemMetricsStorage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		value(rw, r, storage)
	}
}

func value(rw http.ResponseWriter, r *http.Request, storage *repository.MemMetricsStorage) {
	var metName string
	metType := models.MetricType(chi.URLParam(r, "metType"))
	metName = chi.URLParam(r, "metName")

	err := service.ValidateMetricType(metType)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	var message string
	switch metType {
	case models.TypeGauge:
		val, err := storage.GetGauge(metName)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		// message = fmt.Sprintf("%s type of %s: %f", metName, metType, float64(val))
		message = strconv.FormatFloat(float64(val), 'f', -1, 64)

	case models.TypeCounter:
		val, err := storage.GetCounter(metName)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		// message = fmt.Sprintf("%s type of %s: %d", metName, metType, int64(val))
		message = fmt.Sprintf("%d", int64(val))
	}
	RespondTextOK(rw, message)
}

func ValueJSONHandler(storage *repository.MemMetricsStorage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		valueJSON(w, r, storage, logger)
	}
}

func valueJSON(w http.ResponseWriter, r *http.Request, storage *repository.MemMetricsStorage, logger *zap.SugaredLogger) {
	logger.Debug("decoding request")

	var metric models.Metric
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metric); err != nil {
		logger.Debug("cannot decode metric JSON body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := metric.ValidateMeta(); err != nil {
		logger.Debug("invalid metric", zap.Error(err))
		handleServiceError(w, logger, err)
		return
	}

	message := metric
	switch metric.MType {
	case models.TypeGauge:
		val, err := storage.GetGauge(metric.ID)
		if err != nil {
			logger.Debug("cannot GetGauge", zap.Error(err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		message.Value = &val

	case models.TypeCounter:
		val, err := storage.GetCounter(metric.ID)
		if err != nil {
			logger.Debug("cannot GetCounter", zap.Error(err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		message.Delta = &val
	}

	RespondJSONOK(w, message)
}
