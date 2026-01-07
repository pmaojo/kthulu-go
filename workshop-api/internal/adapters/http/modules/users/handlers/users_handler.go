// @kthulu:handler:users
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/example/workshop-api/internal/adapters/http/modules/users/domain"
)

type UsersHandler struct {
	service domain.UsersService
}

func NewUsersHandler(service domain.UsersService) *UsersHandler {
	return &UsersHandler{service: service}
}

// RegisterRoutes registers all routes for users
func (h *UsersHandler) RegisterRoutes(router *mux.Router) {
	sub := router.PathPrefix("/users").Subrouter()
	sub.HandleFunc("", h.List).Methods("GET")
	sub.HandleFunc("", h.Create).Methods("POST")
	sub.HandleFunc("/{id}", h.GetByID).Methods("GET")
	sub.HandleFunc("/{id}", h.Update).Methods("PUT")
	sub.HandleFunc("/{id}", h.Delete).Methods("DELETE")
}

func (h *UsersHandler) Create(w http.ResponseWriter, r *http.Request) {
	var entity domain.Users
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CreateUsers(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

func (h *UsersHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	entity, err := h.service.GetUsersByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

func (h *UsersHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var entity domain.Users
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entity.ID = uint(id)
	if err := h.service.UpdateUsers(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

func (h *UsersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteUsers(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UsersHandler) List(w http.ResponseWriter, r *http.Request) {
	entities, err := h.service.ListUserses()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
