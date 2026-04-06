package handler

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/F3dosik/metalert/internal/service"
)

var errPingFailed = "Database connection failed"

// PingDB возвращает HTTP-хендлер для проверки соединения с базой данных.
//
// Маршрут: GET /ping
//
// Возможные ответы:
//   - 200 OK — соединение с БД установлено
//   - 500 Internal Server Error — БД не инициализирована или недоступна
func PingDB(svc service.MetricsService, logger *zap.SugaredLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := svc.Ping(); err != nil {
			if errors.Is(err, service.ErrPingNotSupported) {
				http.Error(w, "DB not initialized", http.StatusInternalServerError)
				return
			}
			logger.Errorw(errPingFailed, "err", err)
			http.Error(w, errPingFailed, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
