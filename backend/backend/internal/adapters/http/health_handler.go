// @kthulu:core
package adapterhttp

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"backend/core"
)

// HealthHandler provides health check endpoints
type HealthHandler struct {
	db      *sql.DB
	logger  *zap.Logger
	version string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(p struct {
	fx.In
	DB     *sql.DB
	Logger *zap.Logger
	Config *core.Config
}) *HealthHandler {
	return &HealthHandler{
		db:      p.DB,
		logger:  p.Logger,
		version: p.Config.Version,
	}
}

// RegisterRoutes registers health check routes
func (h *HealthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.healthCheck)
	r.Get("/health/ready", h.readinessCheck)
	r.Get("/health/live", h.livenessCheck)
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// healthCheck godoc
// @Summary Comprehensive health check
// @Description Returns the overall health status of the service including database connectivity
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Failure 503 {object} HealthResponse "Service is unhealthy"
// @Router /health [get]
func (h *HealthHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	status := "healthy"

	// Check database connection only if DB is configured
	if h.db == nil {
		checks["database"] = "unavailable"
	} else if err := h.db.Ping(); err != nil {
		h.logger.Error("Database health check failed", zap.Error(err))
		checks["database"] = "unhealthy: " + err.Error()
		status = "unhealthy"
	} else {
		checks["database"] = "healthy"
	}

	response := HealthResponse{
		Status:    status,
		Version:   h.version,
		Timestamp: time.Now(),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	if status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_ = json.NewEncoder(w).Encode(response)
}

// readinessCheck godoc
// @Summary Readiness check
// @Description Indicates if the service is ready to accept traffic
// @Tags Health
// @Produce plain
// @Success 200 {string} string "Ready"
// @Failure 503 {string} string "Service not ready"
// @Router /health/ready [get]
func (h *HealthHandler) readinessCheck(w http.ResponseWriter, r *http.Request) {
	// Check if database is accessible
	if h.db != nil {
		if err := h.db.Ping(); err != nil {
			h.logger.Error("Readiness check failed - database not accessible", zap.Error(err))
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service not ready"))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

// livenessCheck godoc
// @Summary Liveness check
// @Description Indicates if the service is alive
// @Tags Health
// @Produce plain
// @Success 200 {string} string "Alive"
// @Router /health/live [get]
func (h *HealthHandler) livenessCheck(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - if we can respond, we're alive
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alive"))
}
