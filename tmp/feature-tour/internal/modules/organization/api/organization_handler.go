// @kthulu:handler:organization
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"feature-tour/internal/modules/organization/core"
	
)

type OrganizationHandler struct {
	service core.OrganizationService
}

func NewOrganizationHandler(service core.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{service: service}
}

// RegisterRoutes registers all routes for organization
func (h *OrganizationHandler) RegisterRoutes(router *mux.Router) {
	sub := router.PathPrefix("/organization").Subrouter()
	

	sub.HandleFunc("/", h.List).Methods("GET")
	sub.HandleFunc("/", h.Create).Methods("POST")
	sub.HandleFunc("/{id}", h.GetByID).Methods("GET")
	sub.HandleFunc("/{id}", h.Update).Methods("PUT")
	sub.HandleFunc("/{id}", h.Delete).Methods("DELETE")
}

// Create handles the creation of a new organization
// @Summary      Create a new organization
// @Description  Creates a new organization with the provided data
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        input body core.Organization true "Organization object"
// @Success      200  {object}  core.Organization
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /organization [post]
func (h *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var entity core.Organization
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := h.service.CreateOrganization(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a organization by its ID
// @Summary      Get organization
// @Description  Get a organization by its ID
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Organization ID"
	// @Success      200  {object}  core.Organization
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Organization not found"
// @Router       /organization/{id} [get]
func (h *OrganizationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	entity, err := h.service.GetOrganizationByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Update handles the update of a organization
func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var entity core.Organization
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	entity.ID = uint(id)
	if err := h.service.UpdateOrganization(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Delete handles the deletion of a organization
func (h *OrganizationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	if err := h.service.DeleteOrganization(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// List retrieves all organizations
func (h *OrganizationHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := core.SearchFilter{
		Query: query,
		Limit: limit,
		Offset: offset,
	}

	entities, err := h.service.ListOrganizations(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
