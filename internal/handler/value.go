package handler

import (
	"fmt"
	"net/http"
	"strconv"

	models "github.com/F3dosik/metalert.git/internal/model"
	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/internal/service"
	"github.com/go-chi/chi/v5"
)

func ValueHandler(storage *repository.MemStorage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		value(rw, r, storage)
	}
}

func value(rw http.ResponseWriter, r *http.Request, storage *repository.MemStorage) {
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
