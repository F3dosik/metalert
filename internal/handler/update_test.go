package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/F3dosik/metalert.git/internal/audit"
	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/F3dosik/metalert.git/pkg/logger"
)

func TestUpdate(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		url         string
		contentType string
		want        int
	}{
		{
			name:        "POST запрос OK",
			method:      http.MethodPost,
			url:         "/update/gauge/name/40",
			contentType: "text/plain",
			want:        http.StatusOK,
		},
		// {
		// 	name:        "Content-Type != text/plain",
		// 	method:      http.MethodPost,
		// 	url:         "/update/gauge/name/40",
		// 	contentType: "application/json",
		// 	want:        http.StatusBadRequest,
		// },
		{
			name:        "GET запрос",
			method:      http.MethodGet,
			url:         "/update/gauge/name/40",
			contentType: "text/plain",
			want:        http.StatusMethodNotAllowed,
		},
		{
			name:        "Отстутствует часть пути",
			method:      http.MethodPost,
			url:         "/update/",
			contentType: "text/plain",
			want:        http.StatusNotFound,
		},
		{
			name:        "Некорретный тип метрики",
			method:      http.MethodPost,
			url:         "/update/gauges",
			contentType: "text/plain",

			want: http.StatusNotFound,
		},
		{
			name:        "Отсутствует имя метрики",
			method:      http.MethodPost,
			url:         "/update/gauge/",
			contentType: "text/plain",
			want:        http.StatusNotFound,
		},
		{
			name:        "Неккоретное значение(gauge)",
			method:      http.MethodPost,
			url:         "/update/gauge/name/str",
			contentType: "text/plain",
			want:        http.StatusBadRequest,
		},
		{
			name:        "Неккоретое значение(counter)",
			method:      http.MethodPost,
			url:         "/update/counter/name/40.314",
			contentType: "text/plain",
			want:        http.StatusBadRequest,
		},
		{
			name:        "Слишком много параметров",
			method:      http.MethodPost,
			url:         "/update/gauge/name/40/str/str",
			contentType: "text/plain",
			want:        http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewMemMetricsStorage()

			baseLogger, sugarLogger := logger.NewLogger("development")
			defer func() { _ = baseLogger.Sync() }()

			dispatcher := &audit.AuditDispatcher{}

			r := chi.NewRouter()
			r.Route("/update", func(r chi.Router) {
				r.Post("/{metType}/{metName}/{metValue}", UpdateHandler(storage, dispatcher, sugarLogger))
			})

			req := httptest.NewRequest(tt.method, tt.url, nil)
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if rr.Code != tt.want {
				t.Errorf("status = %d, want %d", rr.Code, tt.want)
			}
		})
	}
}
