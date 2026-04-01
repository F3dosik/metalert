package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/F3dosik/metalert/internal/crypto"
	"go.uber.org/zap"
)

// DecryptMiddleware возвращает HTTP middleware, которое прозрачно расшифровывает
// тело запроса перед передачей его следующему обработчику.
//
// Расшифровка выполняется только если выполнены оба условия:
//   - передан ненулевой приватный RSA-ключ;
//   - запрос содержит заголовок X-Encrypted: true.
//
// В случае успеха тело запроса заменяется расшифрованными данными,
// Content-Length обновляется, а Content-Type устанавливается в application/json.
// При ошибке чтения или расшифровки возвращается 400 Bad Request.
func DecryptMiddleware(privateKey *rsa.PrivateKey, logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey == nil || r.Header.Get("X-Encrypted") != "true" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "read body error", http.StatusBadRequest)
				logger.Debugw("decrype middleware: read body", "error", err)
				return
			}
			defer r.Body.Close()

			decrypted, err := crypto.Decrypt(body, privateKey)
			if err != nil {
				http.Error(w, "decrypt error", http.StatusBadRequest)
				logger.Debugw("decrypt middleware: decrypt", "error", err)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.ContentLength = int64(len(decrypted))
			r.Header.Set("Content-Type", "application/json")

			next.ServeHTTP(w, r)
		})
	}
}
