// @kthulu:handler:cars
package handlers

import (
        "encoding/json"
        "net/http"
        "strconv"

        "github.com/gorilla/mux"
        "car-farm-app/internal/adapters/http/modules/cars/domain"

)

type CarsHandler struct {
        service domain.CarsService
}

func NewCarsHandler(service domain.CarsService) *CarsHandler {
        return &CarsHandler{service: service}
}

func (h *CarsHandler) RegisterRoutes(r *mux.Router) {
		sub := r


        sub.HandleFunc("/carss", h.List).Methods("GET")
        sub.HandleFunc("/carss", h.Create).Methods("POST")
        sub.HandleFunc("/carss/{id}", h.GetByID).Methods("GET")
}

// Create handles the creation of a new cars
// @Summary      Create a new cars
// @Description  Creates a new cars with the provided data
// @Tags         carss
// @Accept       json
// @Produce      json
// @Param        input body domain.Cars true "Cars object"
// @Success      200  {object}  domain.Cars
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /carss [post]
func (h *CarsHandler) Create(w http.ResponseWriter, r *http.Request) {
        var entity domain.Cars
        if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }

        if err := h.service.CreateCars(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a cars by its ID
// @Summary      Get cars
// @Description  Get a cars by its ID
// @Tags         carss
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Cars ID"
// @Success      200  {object}  domain.Cars
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Cars not found"
// @Router       /carss/{id} [get]
func (h *CarsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.ParseUint(vars["id"], 10, 32)
        if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        entity, err := h.service.GetCarsByID(uint(id))
        if err != nil {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// List retrieves all carss
// @Summary      List carss
// @Description  Get a list of all carss
// @Tags         carss
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Cars
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /carss [get]
func (h *CarsHandler) List(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		filter := domain.SearchFilter{
			Query: query,
			Limit: limit,
			Offset: offset,
		}

        entities, err := h.service.ListCarss(filter)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entities)
}
