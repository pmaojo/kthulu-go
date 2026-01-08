// @kthulu:module:projects
package adapterhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

// ProjectHandler exposes project management endpoints.
type ProjectHandler struct {
	project *usecase.ProjectUseCase
	log     *zap.SugaredLogger
}

// NewProjectHandler constructs ProjectHandler with required dependencies.
func NewProjectHandler(project *usecase.ProjectUseCase, logger *zap.Logger) *ProjectHandler {
	return &ProjectHandler{
		project: project,
		log:     logger.Sugar(),
	}
}

// RegisterRoutes attaches project routes to the router.
func (h *ProjectHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/v1/projects/plan", h.planProject)
	r.Post("/api/v1/projects", h.generateProject)
	r.Get("/api/v1/projects/{id}", h.getProject)
	r.Patch("/api/v1/projects/{id}", h.updateProject)
	r.Delete("/api/v1/projects/{id}", h.deleteProject)
	r.Get("/api/v1/projects", h.listProjects)
}

// planProject godoc
// @Summary Plan a project
// @Description Creates a project plan based on the provided configuration
// @Tags Projects
// @Accept json
// @Produce json
// @Param request body domain.ProjectRequest true "Project configuration"
// @Success 200 {object} domain.ProjectPlan "Project plan created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/projects/plan [post]
func (h *ProjectHandler) planProject(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	var req domain.ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode plan project request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infow("Plan project request", "name", req.Name)

	plan, err := h.project.PlanProject(r.Context(), req)
	if err != nil {
		logger.Errorw("Failed to plan project", "name", req.Name, "error", err)
		http.Error(w, "Failed to plan project", http.StatusInternalServerError)
		return
	}

	logger.Infow("Project planned successfully", "name", req.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// generateProject godoc
// @Summary Generate a project
// @Description Creates and persists a new project based on the configuration
// @Tags Projects
// @Accept json
// @Produce json
// @Param request body domain.ProjectRequest true "Project configuration"
// @Success 201 {object} domain.Project "Project created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 409 {object} map[string]string "Project already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/projects [post]
func (h *ProjectHandler) generateProject(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	var req domain.ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode generate project request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infow("Generate project request", "name", req.Name)

	project, err := h.project.GenerateProject(r.Context(), req)
	if err != nil {
		logger.Errorw("Failed to generate project", "name", req.Name, "error", err)
		if err.Error() == "project with name "+req.Name+" already exists" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to generate project", http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Project generated successfully", "name", req.Name, "id", project.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

// getProject godoc
// @Summary Get a project
// @Description Retrieves a project by ID
// @Tags Projects
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} domain.Project "Project retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid project ID"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/projects/{id} [get]
func (h *ProjectHandler) getProject(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Errorw("Invalid project ID", "id", idStr, "error", err)
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	logger.Infow("Get project request", "id", id)

	project, err := h.project.GetProject(r.Context(), uint(id))
	if err != nil {
		logger.Errorw("Failed to get project", "id", id, "error", err)
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get project", http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Project retrieved successfully", "id", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// updateProject godoc
// @Summary Update a project
// @Description Updates an existing project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param request body domain.ProjectRequest true "Project update configuration"
// @Success 200 {object} domain.Project "Project updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/projects/{id} [patch]
func (h *ProjectHandler) updateProject(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Errorw("Invalid project ID", "id", idStr, "error", err)
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req domain.ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorw("Failed to decode update project request", "id", id, "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Infow("Update project request", "id", id)

	project, err := h.project.UpdateProject(r.Context(), uint(id), req)
	if err != nil {
		logger.Errorw("Failed to update project", "id", id, "error", err)
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update project", http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Project updated successfully", "id", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// deleteProject godoc
// @Summary Delete a project
// @Description Removes a project by ID
// @Tags Projects
// @Param id path int true "Project ID"
// @Success 204 "Project deleted successfully"
// @Failure 400 {object} map[string]string "Invalid project ID"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/projects/{id} [delete]
func (h *ProjectHandler) deleteProject(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Errorw("Invalid project ID", "id", idStr, "error", err)
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	logger.Infow("Delete project request", "id", id)

	err = h.project.DeleteProject(r.Context(), uint(id))
	if err != nil {
		logger.Errorw("Failed to delete project", "id", id, "error", err)
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		}
		return
	}

	logger.Infow("Project deleted successfully", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

// listProjects godoc
// @Summary List projects
// @Description Retrieves a list of projects with pagination
// @Tags Projects
// @Produce json
// @Param limit query int false "Maximum number of projects to return" default(10)
// @Param offset query int false "Number of projects to skip" default(0)
// @Success 200 {array} domain.Project "Projects retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid query parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/projects [get]
func (h *ProjectHandler) listProjects(w http.ResponseWriter, r *http.Request) {
	logger := h.log

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	logger.Infow("List projects request", "limit", limit, "offset", offset)

	projects, err := h.project.ListProjects(r.Context(), limit, offset)
	if err != nil {
		logger.Errorw("Failed to list projects", "error", err)
		http.Error(w, "Failed to list projects", http.StatusInternalServerError)
		return
	}

	logger.Infow("Projects listed successfully", "count", len(projects))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}
