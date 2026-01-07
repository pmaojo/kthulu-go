// @kthulu:handler:maintenance
package handlers

import (
        "encoding/json"
        "net/http"
        "strconv"

        "github.com/gorilla/mux"
        "car-farm-app/internal/adapters/http/modules/maintenance/domain"

)

type MaintenanceHandler struct {
        service domain.MaintenanceService
}

func NewMaintenanceHandler(service domain.MaintenanceService) *MaintenanceHandler {
        return &MaintenanceHandler{service: service}
}

func (h *MaintenanceHandler) RegisterRoutes(r *mux.Router) {
		sub := r


        sub.HandleFunc("/maintenances", h.List).Methods("GET")
        sub.HandleFunc("/maintenances", h.Create).Methods("POST")
        sub.HandleFunc("/maintenances/{id}", h.GetByID).Methods("GET")
}

// Create handles the creation of a new maintenance
// @Summary      Create a new maintenance
// @Description  Creates a new maintenance with the provided data
// @Tags         maintenances
// @Accept       json
// @Produce      json
// @Param        input body domain.Maintenance true "Maintenance object"
// @Success      200  {object}  domain.Maintenance
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /maintenances [post]
func (h *MaintenanceHandler) Create(w http.ResponseWriter, r *http.Request) {
        var entity domain.Maintenance
        if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }

        if err := h.service.CreateMaintenance(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a maintenance by its ID
// @Summary      Get maintenance
// @Description  Get a maintenance by its ID
// @Tags         maintenances
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Maintenance ID"
// @Success      200  {object}  domain.Maintenance
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Maintenance not found"
// @Router       /maintenances/{id} [get]
func (h *MaintenanceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.ParseUint(vars["id"], 10, 32)
        if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        entity, err := h.service.GetMaintenanceByID(uint(id))
        if err != nil {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// List retrieves all maintenances
// @Summary      List maintenances
// @Description  Get a list of all maintenances
// @Tags         maintenances
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Maintenance
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /maintenances [get]
func (h *MaintenanceHandler) List(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		filter := domain.SearchFilter{
			Query: query,
			Limit: limit,
			Offset: offset,
		}

        entities, err := h.service.ListMaintenances(filter)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entities)
}
