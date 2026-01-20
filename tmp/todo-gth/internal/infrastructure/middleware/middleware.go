// @kthulu:infrastructure:middleware
package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"todo-gth/internal/infrastructure/observability"
)

func ObservabilityMiddleware(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			logger.Info("http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration", duration,
			)

			observability.HTTPRequestDuration.WithLabelValues(
				r.URL.Path,
				r.Method,
				strconv.Itoa(rw.status),
			).Observe(duration.Seconds())
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
