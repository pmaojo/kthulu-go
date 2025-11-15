// @kthulu:core
package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"

	"github.com/pmaojo/kthulu-go/backend/internal/observability"
)

// RecoveryMiddleware creates a middleware that recovers from panics and logs them
func RecoveryMiddleware(logger observability.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get request-specific logger if available
					requestLogger := GetLogger(r.Context())
					if requestLogger == nil {
						requestLogger = logger
					}

					// Capture the panic in Sentry
					var captureErr error
					if e, ok := err.(error); ok {
						captureErr = e
					} else {
						captureErr = fmt.Errorf("%v", err)
					}
					sentry.CaptureException(captureErr)

					// Log the panic with stack trace
					requestLogger.Error("Panic recovered",
						zap.String("request_id", GetRequestID(r.Context())),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.Any("error", err),
						zap.String("stack", string(debug.Stack())),
					)

					// Return 500 error
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// HealthCheckMiddleware provides a simple health check endpoint
func HealthCheckMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" && r.Method == "GET" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"status":"ok","service":"kthulu-api"}`)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
