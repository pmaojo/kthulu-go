// @kthulu:module:templates
package adapterhttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/kthulu/kthulu-go/backend/internal/domain"
	"github.com/kthulu/kthulu-go/backend/internal/usecase"
)

// TemplateHandler exposes template management endpoints.
type TemplateHandler struct {
	template *usecase.TemplateUseCase
	log      *zap.SugaredLogger
}

// NewTemplateHandler constructs TemplateHandler with required dependencies.
func NewTemplateHandler(template *usecase.TemplateUseCase, logger *zap.Logger) *TemplateHandler {
	return &TemplateHandler{
		template: template,
		log:      logger.Sugar(),
	}
}

// RegisterRoutes attaches template routes to the router.
func (h *TemplateHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/templates", h.listTemplates)
	r.Post("/api/v1/templates", h.createTemplate)
	r.Get("/api/v1/templates/{name}", h.getTemplate)
	r.Put("/api/v1/templates/{name}", h.updateTemplate)
	r.Delete("/api/v1/templates/{name}", h.deleteTemplate)
	r.Post("/api/v1/templates/{name}/validate", h.validateTemplate)
	r.Post("/api/v1/templates/render", h.renderTemplate)
	r.Post("/api/v1/templates/registries", h.addRegistry)
	r.Delete("/api/v1/templates/registries/{name}", h.removeRegistry)
}

// listTemplates godoc
// @Summary List templates
// @Description Retrieves a list of available templates
// @Tags Templates
// @Produce json
// @Param limit query int false "Maximum number of templates to return" default(50)
// @Param offset query int false "Number of templates to skip" default(0)
// @Success 200 {array} domain.Template "Templates retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid query parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/templates [get]
func (h *TemplateHandler) listTemplates(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	limit := 50 // default
	offset := 0 // default

	logger.Infow("List templates request", "limit", limit, "offset", offset)

	templates, err := h.template.ListTemplates(r.Context(), limit, offset)
	if err != nil {
		logger.Errorw("Failed to list templates", "error", err)
		http.Error(w, "Failed to list templates", http.StatusInternalServerError)
		return
	}

	logger.Infow("Templates listed successfully", "count", len(templates))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// createTemplate godoc
// @Summary Create a template
// @Description Creates a new template
// @Tags Templates
// @Accept json
// @Produce json
// @Param template body domain.Template true "Template data"
// @Success 201 {object} domain.Template "Template created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 409 {object} map[string]string "Template already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/templates [post]
func (h *TemplateHandler) createTemplate(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	var req domain.Template
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode create template request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infow("Create template request", "name", req.Name)

	template, err := h.template.CreateTemplate(r.Context(), req)
	if err != nil {
		logger.Errorw("Failed to create template", "name", req.Name, "error", err)
		if err.Error() == "template with name "+req.Name+" already exists" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to create template", http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Template created successfully", "name", req.Name, "id", template.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}

// getTemplate godoc
// @Summary Get a template
// @Description Retrieves a specific template by name
// @Tags Templates
// @Produce json
// @Param name path string true "Template name"
// @Success 200 {object} domain.Template "Template retrieved successfully"
// @Failure 404 {object} map[string]string "Template not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/templates/{name} [get]
func (h *TemplateHandler) getTemplate(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	name := chi.URLParam(r, "name")

	logger.Infow("Get template request", "name", name)

	template, err := h.template.GetTemplate(r.Context(), name)
	if err != nil {
		logger.Errorw("Failed to get template", "name", name, "error", err)
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	logger.Infow("Template retrieved successfully", "name", name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// updateTemplate godoc
// @Summary Update a template
// @Description Updates an existing template
// @Tags Templates
// @Accept json
// @Produce json
// @Param name path string true "Template name"
// @Param template body domain.Template true "Template update data"
// @Success 200 {object} domain.Template "Template updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Template not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/templates/{name} [put]
func (h *TemplateHandler) updateTemplate(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	name := chi.URLParam(r, "name")

	var req domain.Template
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode update template request", "name", name, "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infow("Update template request", "name", name)

	template, err := h.template.UpdateTemplate(r.Context(), name, req)
	if err != nil {
		logger.Errorw("Failed to update template", "name", name, "error", err)
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	logger.Infow("Template updated successfully", "name", name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// deleteTemplate godoc
// @Summary Delete a template
// @Description Removes a template by name
// @Tags Templates
// @Param name path string true "Template name"
// @Success 204 "Template deleted successfully"
// @Failure 404 {object} map[string]string "Template not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/templates/{name} [delete]
func (h *TemplateHandler) deleteTemplate(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	name := chi.URLParam(r, "name")

	logger.Infow("Delete template request", "name", name)

	err := h.template.DeleteTemplate(r.Context(), name)
	if err != nil {
		logger.Errorw("Failed to delete template", "name", name, "error", err)
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	logger.Infow("Template deleted successfully", "name", name)
	w.WriteHeader(http.StatusNoContent)
}

// validateTemplate godoc
// @Summary Validate a template
// @Description Validates a template's structure and content
// @Tags Templates
// @Param name path string true "Template name"
// @Success 200 {object} map[string]string "Template is valid"
// @Failure 404 {object} map[string]string "Template not found"
// @Failure 422 {object} map[string]string "Template validation failed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/templates/{name}/validate [post]
func (h *TemplateHandler) validateTemplate(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	name := chi.URLParam(r, "name")

	logger.Infow("Validate template request", "name", name)

	err := h.template.ValidateTemplate(r.Context(), name)
	if err != nil {
		logger.Errorw("Failed to validate template", "name", name, "error", err)
		http.Error(w, "Template validation failed", http.StatusUnprocessableEntity)
		return
	}

	logger.Infow("Template validated successfully", "name", name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
}

// renderTemplate godoc
// @Summary Render a template
// @Description Renders a template with provided variables
// @Tags Templates
// @Accept json
// @Produce json
// @Param request body domain.TemplateRenderRequest true "Render request"
// @Success 200 {object} domain.TemplateRenderResult "Template rendered successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Template not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/templates/render [post]
func (h *TemplateHandler) renderTemplate(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	var req domain.TemplateRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode render template request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infow("Render template request", "name", req.Name)

	result, err := h.template.RenderTemplate(r.Context(), req)
	if err != nil {
		logger.Errorw("Failed to render template", "name", req.Name, "error", err)
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	logger.Infow("Template rendered successfully", "name", req.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// addRegistry godoc
// @Summary Add template registry
// @Description Adds a new template registry source
// @Tags Templates
// @Accept json
// @Produce json
// @Param request body map[string]string true "Registry data"
// @Success 201 {object} domain.TemplateRegistry "Registry added successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 409 {object} map[string]string "Registry already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/templates/registries [post]
func (h *TemplateHandler) addRegistry(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	var req struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode add registry request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infow("Add registry request", "name", req.Name, "url", req.URL)

	registry, err := h.template.AddRegistry(r.Context(), req.Name, req.URL)
	if err != nil {
		logger.Errorw("Failed to add registry", "name", req.Name, "error", err)
		if err.Error() == "registry with name "+req.Name+" already exists" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to add registry", http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Registry added successfully", "name", req.Name, "id", registry.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(registry)
}

// removeRegistry godoc
// @Summary Remove template registry
// @Description Removes a template registry source
// @Tags Templates
// @Param name path string true "Registry name"
// @Success 204 "Registry removed successfully"
// @Failure 404 {object} map[string]string "Registry not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/templates/registries/{name} [delete]
func (h *TemplateHandler) removeRegistry(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	name := chi.URLParam(r, "name")

	logger.Infow("Remove registry request", "name", name)

	err := h.template.RemoveRegistry(r.Context(), name)
	if err != nil {
		logger.Errorw("Failed to remove registry", "name", name, "error", err)
		http.Error(w, "Registry not found", http.StatusNotFound)
		return
	}

	logger.Infow("Registry removed successfully", "name", name)
	w.WriteHeader(http.StatusNoContent)
}
