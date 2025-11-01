// Package handler содержит HTTP-хэндлеры сервера метрик.
// Здесь реализованы функции обработки POST-запросов,
// проверки URL, типов и значений метрик и возврата ответов клиенту.
package handler

import (
	"context"
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
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		logger.Debug("cannot decode metric JSON body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := updateMetrics(r.Context(), storage, []models.Metric{metric}); err != nil {
		handleServiceError(w, logger, err)
		return
	}

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

func UpdatesJSONHandler(storage repository.MetricsStorage, logger *zap.SugaredLogger, asyncSave bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updatesJSON(w, r, storage, logger, asyncSave)
	}
}

func updatesJSON(w http.ResponseWriter, r *http.Request, storage repository.MetricsStorage, logger *zap.SugaredLogger, saveOnUpdate bool) {
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
