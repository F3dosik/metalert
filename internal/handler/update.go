// Package handler содержит HTTP-хэндлеры сервера метрик.
// Здесь реализованы функции обработки POST-запросов,
// проверки URL, типов и значений метрик и возврата ответов клиенту.
package handler

import (
	"errors"
	"fmt"
	"log"

	"net/http"

	models "github.com/F3dosik/metalert.git/internal/model"
	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/internal/service"
	"github.com/go-chi/chi/v5"
)

func UpdateHandler(storage *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		update(w, r, storage)
	}
}

func update(rw http.ResponseWriter, r *http.Request, storage *repository.MemStorage) {
	// if !isPlainText(r) {
	// 	http.Error(rw, errInvalidContentType.Error(), http.StatusBadRequest)
	// 	return
	// }

	var metName, metValue string

	metType := models.MetricType(chi.URLParam(r, "metType"))
	metName = chi.URLParam(r, "metName")
	metValue = chi.URLParam(r, "metValue")

	val, err := service.CheckAndParseValue(metType, metName, metValue)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidType):
			http.Error(rw, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrNoName):
			http.Error(rw, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrInvalidVal):
			http.Error(rw, err.Error(), http.StatusBadRequest)
		default:
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	switch v := val.(type) {
	case models.Gauge:
		storage.SetGauge(metName, v)
	case models.Counter:
		storage.UpdateCounter(metName, v)
	}

	message := fmt.Sprintf("Метрика %s успешно обновлена\r\n", metName)
	log.Print(message)
	RespondTextOK(rw, message)

}
