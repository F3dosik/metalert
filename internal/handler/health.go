package handler

import (
	"net/http"

	"github.com/F3dosik/metalert.git/internal/repository"
	"go.uber.org/zap"
)

var (
	errPingFailed = "Database connection failed"
)

func PingDB(storage repository.MetricsStorage, useDB bool, logger *zap.SugaredLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if useDB {
			dbStorage, ok := storage.(*repository.DBMetricStorage)
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
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No DB"))
	})
}
