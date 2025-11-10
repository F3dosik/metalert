package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"
)

var errIncorrectSignature = errors.New("подпись запроса неверна")

func VerifySignature(key string, logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dataHex := r.Header.Get("HashSHA256")
			if key == "" || dataHex == "" {
				next.ServeHTTP(w, r)
				return
			}

			data, err := hex.DecodeString(dataHex)
			if err != nil {
				logger.Errorf("hex decode err: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Errorf("read body err: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			r.Body = io.NopCloser(bytes.NewBuffer(body))

			h := hmac.New(sha256.New, []byte(key))
			h.Write(body)
			sign := h.Sum(nil)

			if !hmac.Equal(sign, data) {
				logger.Debug(errIncorrectSignature)
				http.Error(w, errIncorrectSignature.Error(), http.StatusBadRequest)
				return
			}

			logger.Debug("подпись запроса подлинная")
			next.ServeHTTP(w, r)
		})
	}
}
