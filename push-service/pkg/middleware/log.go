package middleware

import (
	"net/http"
	"time"

	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)

		logger.Info("HTTP request", logger.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": duration.Milliseconds(),
			"remote_addr": r.RemoteAddr,
		})
	})
}
