// @kthulu:module:modules
package adapterhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"backend/internal/domain"
	"backend/internal/usecase"
)

// ModuleHandler exposes module catalog endpoints.
type ModuleHandler struct {
	module *usecase.ModuleUseCase
	log    *zap.SugaredLogger
}

// NewModuleHandler constructs ModuleHandler with required dependencies.
func NewModuleHandler(module *usecase.ModuleUseCase, logger *zap.Logger) *ModuleHandler {
	return &ModuleHandler{
		module: module,
		log:    logger.Sugar(),
	}
}

// RegisterRoutes attaches module routes to the router.
func (h *ModuleHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/modules", h.listModules)
	r.Get("/api/v1/modules/{name}", h.getModule)
	r.Post("/api/v1/modules/validate", h.validateModules)
	r.Post("/api/v1/modules/plan", h.planModules)
}

// listModules godoc
// @Summary List modules
// @Description Retrieves a list of available modules, optionally filtered by category
// @Tags Modules
// @Produce json
// @Param category query string false "Filter by category"
// @Param limit query int false "Maximum number of modules to return" default(50)
// @Param offset query int false "Number of modules to skip" default(0)
// @Success 200 {array} domain.ModuleInfo "Modules retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid query parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/modules [get]
func (h *ModuleHandler) listModules(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	category := r.URL.Query().Get("category")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	logger.Infow("List modules request", "category", category, "limit", limit, "offset", offset)

	modules, err := h.module.ListModules(r.Context(), categoryPtr, limit, offset)
	if err != nil {
		logger.Errorw("Failed to list modules", "error", err)
		http.Error(w, "Failed to list modules", http.StatusInternalServerError)
		return
	}

	logger.Infow("Modules listed successfully", "count", len(modules))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modules)
}

// getModule godoc
// @Summary Get a module
// @Description Retrieves a specific module by name
// @Tags Modules
// @Produce json
// @Param name path string true "Module name"
// @Success 200 {object} domain.ModuleInfo "Module retrieved successfully"
// @Failure 404 {object} map[string]string "Module not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/modules/{name} [get]
func (h *ModuleHandler) getModule(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	name := chi.URLParam(r, "name")

	logger.Infow("Get module request", "name", name)

	module, err := h.module.GetModule(r.Context(), name)
	if err != nil {
		logger.Errorw("Failed to get module", "name", name, "error", err)
		if err == domain.ErrModuleNotFound {
			http.Error(w, "Module not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get module", http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Module retrieved successfully", "name", name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(module)
}

// validateModules godoc
// @Summary Validate modules
// @Description Validates a set of modules for compatibility and dependencies
// @Tags Modules
// @Accept json
// @Produce json
// @Param request body map[string][]string true "Module names to validate"
// @Success 200 {object} domain.ModuleValidationResult "Validation result"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/modules/validate [post]
func (h *ModuleHandler) validateModules(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	var req struct {
		Modules []string `json:"modules"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode validate modules request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infow("Validate modules request", "modules", req.Modules)

	result, err := h.module.ValidateModules(r.Context(), req.Modules)
	if err != nil {
		logger.Errorw("Failed to validate modules", "error", err)
		http.Error(w, "Failed to validate modules", http.StatusInternalServerError)
		return
	}

	logger.Infow("Modules validated successfully", "valid", result.Valid)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// planModules godoc
// @Summary Plan module injection
// @Description Creates an injection plan for the given modules
// @Tags Modules
// @Accept json
// @Produce json
// @Param request body map[string][]string true "Module names to plan"
// @Success 200 {object} domain.ModuleInjectionPlan "Injection plan"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/modules/plan [post]
func (h *ModuleHandler) planModules(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	var req struct {
		Modules []string `json:"modules"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode plan modules request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infow("Plan modules request", "modules", req.Modules)

	plan, err := h.module.PlanModules(r.Context(), req.Modules)
	if err != nil {
		logger.Errorw("Failed to plan modules", "error", err)
		http.Error(w, "Failed to plan modules", http.StatusInternalServerError)
		return
	}

	logger.Infow("Modules planned successfully", "modules", len(plan.InjectedModules))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}
