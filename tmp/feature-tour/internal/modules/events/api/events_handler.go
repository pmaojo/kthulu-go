// @kthulu:handler:events
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"feature-tour/internal/modules/events/core"
	
)

type EventHandler struct {
	service core.EventService
}

func NewEventHandler(service core.EventService) *EventHandler {
	return &EventHandler{service: service}
}

// RegisterRoutes registers all routes for events
func (h *EventHandler) RegisterRoutes(router *mux.Router) {
	sub := router.PathPrefix("/events").Subrouter()
	

	sub.HandleFunc("/", h.List).Methods("GET")
	sub.HandleFunc("/", h.Create).Methods("POST")
	sub.HandleFunc("/{id}", h.GetByID).Methods("GET")
	sub.HandleFunc("/{id}", h.Update).Methods("PUT")
	sub.HandleFunc("/{id}", h.Delete).Methods("DELETE")
}

// Create handles the creation of a new events
// @Summary      Create a new events
// @Description  Creates a new events with the provided data
// @Tags         Events
// @Accept       json
// @Produce      json
// @Param        input body core.Event true "Event object"
// @Success      200  {object}  core.Event
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /events [post]
func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	var entity core.Event
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := h.service.CreateEvent(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a events by its ID
// @Summary      Get events
// @Description  Get a events by its ID
// @Tags         Events
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Event ID"
	// @Success      200  {object}  core.Event
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Event not found"
// @Router       /events/{id} [get]
func (h *EventHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	entity, err := h.service.GetEventByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Update handles the update of a events
func (h *EventHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var entity core.Event
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	entity.ID = uint(id)
	if err := h.service.UpdateEvent(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Delete handles the deletion of a events
func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	if err := h.service.DeleteEvent(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// List retrieves all eventss
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := core.SearchFilter{
		Query: query,
		Limit: limit,
		Offset: offset,
	}

	entities, err := h.service.ListEvents(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
