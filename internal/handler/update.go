// Package handler содержит HTTP-хэндлеры сервера метрик.
// Здесь реализованы функции обработки POST-запросов,
// проверки URL, типов и значений метрик и возврата ответов клиенту.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	// "log"
	"net/http"
	"strings"

	models "github.com/F3dosik/metalert.git/internal/model"
	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/internal/service"
)

var (
	errInvalidURL         = errors.New("error: invalid URL: expected format /update/<TYPE>/<NAME>/<VALUE>")
	errInvalidMethod      = errors.New("error: only POST request are allowed")
	errInvalidContentType = errors.New("error: Content-Type should be text/plain")
)

const ContentType = "text/plain"

func UpdateHandler(storage *repository.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		update(w, r, storage)
	}
}

func update(w http.ResponseWriter, r *http.Request, storage *repository.MemStorage) {
	if r.Method != http.MethodPost {
		http.Error(w, errInvalidMethod.Error(), http.StatusMethodNotAllowed)
		return
	}

	if !isPlainText(r) {
		http.Error(w, errInvalidContentType.Error(), http.StatusBadRequest)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/update/")
	parts := strings.Split(path, "/")

	var metType, metName, metValue string

	switch len(parts) {
	case 1:
		metType = parts[0]
	case 2:
		metType, metName = parts[0], parts[1]
	case 3:
		metType, metName, metValue = parts[0], parts[1], parts[2]
	default:
		http.Error(w, errInvalidURL.Error(), http.StatusBadRequest)
		return
	}
	// log.Printf("Принята метрика %s %s = %s", metType, metName, metValue)
	val, err := service.CheckURL(metType, metName, metValue)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidType):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrNoName):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, service.ErrInvalidVal):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	switch v := val.(type) {
	case models.Gauge:
		storage.UpdateGauge(metName, v)
	case models.Counter:
		storage.UpdateCounter(metName, v)
	}

	RespondOKText(w, metName)

	// fmt.Println(storage.Gauges)
	// fmt.Println(storage.Counters)
}

func isPlainText(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, ContentType)
}

type UpdateResponse struct {
	Status string `json:"status"`
	Metric string `json:"metric"`
}

func RespondOKJSON(w http.ResponseWriter, metric string) {
	resp := UpdateResponse{
		Status: "ok",
		Metric: metric,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)

}

func RespondOKText(w http.ResponseWriter, metric string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "success: metric %s updated\r\n", metric)
}
