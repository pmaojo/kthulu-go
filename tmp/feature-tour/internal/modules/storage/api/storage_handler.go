// @kthulu:handler:storage
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"feature-tour/internal/modules/storage/core"
	
)

type StorageHandler struct {
	service core.StorageService
}

func NewStorageHandler(service core.StorageService) *StorageHandler {
	return &StorageHandler{service: service}
}

// RegisterRoutes registers all routes for storage
func (h *StorageHandler) RegisterRoutes(router *mux.Router) {
	sub := router.PathPrefix("/storage").Subrouter()
	

	sub.HandleFunc("/", h.List).Methods("GET")
	sub.HandleFunc("/", h.Create).Methods("POST")
	sub.HandleFunc("/{id}", h.GetByID).Methods("GET")
	sub.HandleFunc("/{id}", h.Update).Methods("PUT")
	sub.HandleFunc("/{id}", h.Delete).Methods("DELETE")
}

// Create handles the creation of a new storage
// @Summary      Create a new storage
// @Description  Creates a new storage with the provided data
// @Tags         Storages
// @Accept       json
// @Produce      json
// @Param        input body core.Storage true "Storage object"
// @Success      200  {object}  core.Storage
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /storage [post]
func (h *StorageHandler) Create(w http.ResponseWriter, r *http.Request) {
	var entity core.Storage
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := h.service.CreateStorage(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a storage by its ID
// @Summary      Get storage
// @Description  Get a storage by its ID
// @Tags         Storages
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Storage ID"
	// @Success      200  {object}  core.Storage
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Storage not found"
// @Router       /storage/{id} [get]
func (h *StorageHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	entity, err := h.service.GetStorageByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Update handles the update of a storage
func (h *StorageHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var entity core.Storage
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	entity.ID = uint(id)
	if err := h.service.UpdateStorage(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Delete handles the deletion of a storage
func (h *StorageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	if err := h.service.DeleteStorage(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// List retrieves all storages
func (h *StorageHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := core.SearchFilter{
		Query: query,
		Limit: limit,
		Offset: offset,
	}

	entities, err := h.service.ListStorages(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
