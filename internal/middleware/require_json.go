package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

const (
	ErrUnnsuportedMethod = "Content-Type must be application/json"
)

func RequireJSON(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") != "application/json" {
				logger.Debug("wrong content-type", "got", r.Header.Get("Content-Type"))
				http.Error(w, ErrUnnsuportedMethod, http.StatusMethodNotAllowed)
				return 
			}
			next.ServeHTTP(w, r)
		})
	}
}
