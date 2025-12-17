package middleware

import (
	"net/http"
	"time"

	"github.com/an3wers/notification-serv/internal/pkg/logger"
	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func Logger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UTC()
			_ = start

			wrapped := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			log.With(
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", wrapped.status),
				zap.Int64("duration_ms", time.Since(start).Milliseconds()),
				zap.Int("size", wrapped.size),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()))

			next.ServeHTTP(wrapped, r)

			log.Info("HTTP Request")

		})
	}
}
