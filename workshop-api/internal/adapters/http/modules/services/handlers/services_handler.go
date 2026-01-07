// @kthulu:handler:services
package handlers

import (
        "encoding/json"
        "net/http"
        "strconv"

        "github.com/gorilla/mux"
        "github.com/example/workshop-api/internal/adapters/http/modules/services/domain"
)

type ServicesHandler struct {
        service domain.ServicesService
}

func NewServicesHandler(service domain.ServicesService) *ServicesHandler {
        return &ServicesHandler{service: service}
}

func (h *ServicesHandler) RegisterRoutes(r *mux.Router) {
        r.HandleFunc("/servicess", h.List).Methods("GET")
        r.HandleFunc("/servicess", h.Create).Methods("POST")
        r.HandleFunc("/servicess/{id}", h.GetByID).Methods("GET")
}

// Create handles the creation of a new services
// @Summary      Create a new services
// @Description  Creates a new services with the provided data
// @Tags         servicess
// @Accept       json
// @Produce      json
// @Param        input body domain.Services true "Services object"
// @Success      200  {object}  domain.Services
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /servicess [post]
func (h *ServicesHandler) Create(w http.ResponseWriter, r *http.Request) {
        var entity domain.Services
        if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }

        if err := h.service.CreateServices(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a services by its ID
// @Summary      Get services
// @Description  Get a services by its ID
// @Tags         servicess
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Services ID"
// @Success      200  {object}  domain.Services
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Services not found"
// @Router       /servicess/{id} [get]
func (h *ServicesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.ParseUint(vars["id"], 10, 32)
        if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        entity, err := h.service.GetServicesByID(uint(id))
        if err != nil {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// List retrieves all servicess
// @Summary      List servicess
// @Description  Get a list of all servicess
// @Tags         servicess
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Services
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /servicess [get]
func (h *ServicesHandler) List(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		filter := domain.SearchFilter{
			Query: query,
			Limit: limit,
			Offset: offset,
		}

        entities, err := h.service.ListServicess(filter)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entities)
}
