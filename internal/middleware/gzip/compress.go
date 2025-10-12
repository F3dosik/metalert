package gzip

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func WithCompression(logger *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportGzip := strings.Contains(acceptEncoding, "gzip")
			if supportGzip {
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				cw := newCompressWriter(w)
				ow = cw
				defer func() {
					if err := cw.Close(); err != nil {
						logger.Warnf("Ошибка при закрытии gzip.Writer: %v", err)
					}
				}()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					logger.Errorf("Ошибка при создании gzip.Reader: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer func() {
					if err := cr.Close(); err != nil {
						logger.Warnf("Ошибка при закрытии gzip.Reader: %v", err)
					}
				}()
			}
			next.ServeHTTP(ow, r)
		})
	}
}
