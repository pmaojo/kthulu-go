// @kthulu:handler:farms
package handlers

import (
        "encoding/json"
        "net/http"
        "strconv"

        "github.com/gorilla/mux"
        "car-farm-app/internal/adapters/http/modules/farms/domain"

)

type FarmsHandler struct {
        service domain.FarmsService
}

func NewFarmsHandler(service domain.FarmsService) *FarmsHandler {
        return &FarmsHandler{service: service}
}

func (h *FarmsHandler) RegisterRoutes(r *mux.Router) {
		sub := r


        sub.HandleFunc("/farmss", h.List).Methods("GET")
        sub.HandleFunc("/farmss", h.Create).Methods("POST")
        sub.HandleFunc("/farmss/{id}", h.GetByID).Methods("GET")
}

// Create handles the creation of a new farms
// @Summary      Create a new farms
// @Description  Creates a new farms with the provided data
// @Tags         farmss
// @Accept       json
// @Produce      json
// @Param        input body domain.Farms true "Farms object"
// @Success      200  {object}  domain.Farms
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /farmss [post]
func (h *FarmsHandler) Create(w http.ResponseWriter, r *http.Request) {
        var entity domain.Farms
        if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }

        if err := h.service.CreateFarms(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a farms by its ID
// @Summary      Get farms
// @Description  Get a farms by its ID
// @Tags         farmss
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Farms ID"
// @Success      200  {object}  domain.Farms
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Farms not found"
// @Router       /farmss/{id} [get]
func (h *FarmsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.ParseUint(vars["id"], 10, 32)
        if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        entity, err := h.service.GetFarmsByID(uint(id))
        if err != nil {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// List retrieves all farmss
// @Summary      List farmss
// @Description  Get a list of all farmss
// @Tags         farmss
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Farms
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /farmss [get]
func (h *FarmsHandler) List(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		filter := domain.SearchFilter{
			Query: query,
			Limit: limit,
			Offset: offset,
		}

        entities, err := h.service.ListFarmss(filter)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entities)
}
