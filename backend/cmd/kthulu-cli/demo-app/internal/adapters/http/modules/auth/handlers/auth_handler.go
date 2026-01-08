// @kthulu:handler:auth
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"demo-app/internal/adapters/http/modules/auth/domain"
	
)

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// RegisterRoutes registers all routes for auth
func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	sub := router.PathPrefix("auth").Subrouter()
	

	sub.HandleFunc("/auth", h.List).Methods("GET")
	sub.HandleFunc("/auth", h.Create).Methods("POST")
	sub.HandleFunc("/auth/{id}", h.GetByID).Methods("GET")
	sub.HandleFunc("/auth/{id}", h.Update).Methods("PUT")
	sub.HandleFunc("/auth/{id}", h.Delete).Methods("DELETE")
}

// Create handles the creation of a new auth
// @Summary      Create a new auth
// @Description  Creates a new auth with the provided data
// @Tags         Auths
// @Accept       json
// @Produce      json
// @Param        input body domain.Auth true "Auth object"
// @Success      200  {object}  domain.Auth
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /auth [post]
func (h *AuthHandler) Create(w http.ResponseWriter, r *http.Request) {
	var entity domain.Auth
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := h.service.CreateAuth(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a auth by its ID
// @Summary      Get auth
// @Description  Get a auth by its ID
// @Tags         Auths
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Auth ID"
// @Success      200  {object}  domain.Auth
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Auth not found"
// @Router       /auth/{id} [get]
func (h *AuthHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	entity, err := h.service.GetAuthByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Update handles the update of a auth
func (h *AuthHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var entity domain.Auth
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	entity.ID = uint(id)
	if err := h.service.UpdateAuth(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Delete handles the deletion of a auth
func (h *AuthHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	if err := h.service.DeleteAuth(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// List retrieves all auths
func (h *AuthHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := domain.SearchFilter{
		Query: query,
		Limit: limit,
		Offset: offset,
	}

	entities, err := h.service.ListAuths(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
