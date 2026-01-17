// @kthulu:infrastructure:middleware
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"
)

func RecoveryMiddleware(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					// This specific message "Panic recovered" is monitored by 'kthulu dev'
					// to trigger self-healing logic.
					logger.Error("Panic recovered",
						"error", fmt.Sprintf("%v", err),
						"stack", string(debug.Stack()),
						"path", r.URL.Path,
						"method", r.Method,
					)

					// Return 500 error
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
