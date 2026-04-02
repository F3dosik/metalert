package handler

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/F3dosik/metalert/internal/repository"
)

var (
	errPingFailed = "Database connection failed"
)

// PingDB возвращает HTTP-хендлер для проверки соединения с базой данных.
//
// Маршрут: GET /ping
//
// Хендлер проверяет, что текущее хранилище является DBMetricsStorage
// и успешно отвечает на Ping.
//
// Возможные ответы:
//   - 200 OK — соединение с БД установлено
//   - 500 Internal Server Error — БД не инициализирована или недоступна
func PingDB(storage repository.MetricsStorage, logger *zap.SugaredLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dbStorage, ok := storage.(*repository.DBMetricsStorage)
		if !ok {
			http.Error(w, "DB not initialized", http.StatusInternalServerError)
			return
		}

		if err := dbStorage.Ping(); err != nil {
			logger.Errorw(errPingFailed, "err", err)
			http.Error(w, errPingFailed, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
