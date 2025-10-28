package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/internal/service"
	"github.com/F3dosik/metalert.git/pkg/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func ValueHandler(storage repository.MetricsStorage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		value(rw, r, storage)
	}
}

func value(rw http.ResponseWriter, r *http.Request, storage repository.MetricsStorage) {
	var metName string
	ctx := r.Context()
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
		val, err := storage.GetGauge(ctx, metName)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		message = strconv.FormatFloat(float64(val), 'f', -1, 64)

	case models.TypeCounter:
		val, err := storage.GetCounter(ctx, metName)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		message = strconv.Itoa(int(val))
	}
	RespondTextOK(rw, message)
}

func ValueJSONHandler(storage repository.MetricsStorage, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		valueJSON(w, r, storage, logger)
	}
}

func valueJSON(w http.ResponseWriter, r *http.Request, storage repository.MetricsStorage, logger *zap.SugaredLogger) {
	logger.Debug("decoding request")

	var metric models.Metric
	ctx := r.Context()
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
		val, err := storage.GetGauge(ctx, metric.ID)
		if err != nil {
			logger.Debug("cannot GetGauge", zap.Error(err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		message.Value = &val

	case models.TypeCounter:
		val, err := storage.GetCounter(ctx, metric.ID)
		if err != nil {
			logger.Debug("cannot GetCounter", zap.Error(err))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		message.Delta = &val
	}

	RespondJSONOK(w, message)
}
