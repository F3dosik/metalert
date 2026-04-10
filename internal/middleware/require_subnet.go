package middleware

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

func RequireTrustedSubnet(logger *zap.SugaredLogger, ipNet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ipNet == nil {
				next.ServeHTTP(w, r)
				return
			}
			ipStr := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(ipStr)
			if !ipNet.Contains(ip) {
				logger.Debugw("require trusted subnet: agent IP not in trusted subnet", "ip", ipStr)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
