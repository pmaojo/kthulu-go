// @kthulu:core
package middleware

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/kthulu/kthulu-go/backend/core"
	"github.com/kthulu/kthulu-go/backend/internal/observability"
)

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version,omitempty"`
	Checks    map[string]CheckResult `json:"checks"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
}

// HealthCheckHandler creates a comprehensive health check handler
func HealthCheckHandler(db *sql.DB, logger observability.Logger, version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			return
		}

		start := time.Now()
		checks := make(map[string]CheckResult)
		overallStatus := "healthy"

		// Database health check
		dbStart := time.Now()
		dbErr := core.HealthCheck(db)
		dbLatency := time.Since(dbStart)

		if dbErr != nil {
			checks["database"] = CheckResult{
				Status:  "unhealthy",
				Message: dbErr.Error(),
				Latency: dbLatency,
			}
			overallStatus = "unhealthy"
			logger.Error("Database health check failed", zap.Error(dbErr))
		} else {
			checks["database"] = CheckResult{
				Status:  "healthy",
				Latency: dbLatency,
			}
		}

		// Memory/System checks could be added here
		checks["system"] = CheckResult{
			Status: "healthy",
		}

		// Create health status response
		health := HealthStatus{
			Status:    overallStatus,
			Timestamp: time.Now(),
			Service:   "kthulu-api",
			Version:   version,
			Checks:    checks,
		}

		// Set appropriate HTTP status code
		statusCode := http.StatusOK
		if overallStatus == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		}

		// Log the health check
		totalLatency := time.Since(start)
		logger.Debug("Health check completed",
			zap.String("status", overallStatus),
			zap.Duration("total_latency", totalLatency),
			zap.Int("status_code", statusCode),
		)

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(health); err != nil {
			logger.Error("Failed to encode health response",
				zap.Error(err),
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

// ReadinessCheckHandler creates a readiness check handler (simpler than health check)
func ReadinessCheckHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/readyz" {
			return
		}

		// Quick database ping
		if err := core.HealthCheck(db); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	}
}

// LivenessCheckHandler creates a liveness check handler (very basic)
func LivenessCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/livez" {
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Alive"))
	}
}

// AdvancedHealthMiddleware creates a middleware that handles multiple health endpoints
func AdvancedHealthMiddleware(db *sql.DB, logger observability.Logger, version string) func(next http.Handler) http.Handler {
	healthHandler := HealthCheckHandler(db, logger, version)
	readinessHandler := ReadinessCheckHandler(db)
	livenessHandler := LivenessCheckHandler()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/healthz":
				healthHandler(w, r)
				return
			case "/readyz":
				readinessHandler(w, r)
				return
			case "/livez":
				livenessHandler(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
