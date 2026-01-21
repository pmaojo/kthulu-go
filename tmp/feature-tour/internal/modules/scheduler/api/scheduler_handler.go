// @kthulu:handler:scheduler
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"feature-tour/internal/modules/scheduler/core"
	
)

type SchedulerHandler struct {
	service core.SchedulerService
}

func NewSchedulerHandler(service core.SchedulerService) *SchedulerHandler {
	return &SchedulerHandler{service: service}
}

// RegisterRoutes registers all routes for scheduler
func (h *SchedulerHandler) RegisterRoutes(router *mux.Router) {
	sub := router.PathPrefix("/scheduler").Subrouter()
	

	sub.HandleFunc("/", h.List).Methods("GET")
	sub.HandleFunc("/", h.Create).Methods("POST")
	sub.HandleFunc("/{id}", h.GetByID).Methods("GET")
	sub.HandleFunc("/{id}", h.Update).Methods("PUT")
	sub.HandleFunc("/{id}", h.Delete).Methods("DELETE")
}

// Create handles the creation of a new scheduler
// @Summary      Create a new scheduler
// @Description  Creates a new scheduler with the provided data
// @Tags         Schedulers
// @Accept       json
// @Produce      json
// @Param        input body core.Scheduler true "Scheduler object"
// @Success      200  {object}  core.Scheduler
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /scheduler [post]
func (h *SchedulerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var entity core.Scheduler
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := h.service.CreateScheduler(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a scheduler by its ID
// @Summary      Get scheduler
// @Description  Get a scheduler by its ID
// @Tags         Schedulers
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Scheduler ID"
	// @Success      200  {object}  core.Scheduler
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Scheduler not found"
// @Router       /scheduler/{id} [get]
func (h *SchedulerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	entity, err := h.service.GetSchedulerByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Update handles the update of a scheduler
func (h *SchedulerHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var entity core.Scheduler
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	entity.ID = uint(id)
	if err := h.service.UpdateScheduler(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Delete handles the deletion of a scheduler
func (h *SchedulerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	if err := h.service.DeleteScheduler(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// List retrieves all schedulers
func (h *SchedulerHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := core.SearchFilter{
		Query: query,
		Limit: limit,
		Offset: offset,
	}

	entities, err := h.service.ListSchedulers(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
