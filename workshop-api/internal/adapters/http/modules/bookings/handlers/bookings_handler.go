// @kthulu:handler:bookings
package handlers

import (
        "encoding/json"
        "net/http"
        "strconv"

        "github.com/gorilla/mux"
        "github.com/example/workshop-api/internal/adapters/http/modules/bookings/domain"
)

type BookingsHandler struct {
        service domain.BookingsService
}

func NewBookingsHandler(service domain.BookingsService) *BookingsHandler {
        return &BookingsHandler{service: service}
}

func (h *BookingsHandler) RegisterRoutes(r *mux.Router) {
        r.HandleFunc("/bookingss", h.List).Methods("GET")
        r.HandleFunc("/bookingss", h.Create).Methods("POST")
        r.HandleFunc("/bookingss/{id}", h.GetByID).Methods("GET")
}

// Create handles the creation of a new bookings
// @Summary      Create a new bookings
// @Description  Creates a new bookings with the provided data
// @Tags         bookingss
// @Accept       json
// @Produce      json
// @Param        input body domain.Bookings true "Bookings object"
// @Success      200  {object}  domain.Bookings
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /bookingss [post]
func (h *BookingsHandler) Create(w http.ResponseWriter, r *http.Request) {
        var entity domain.Bookings
        if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }

        if err := h.service.CreateBookings(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a bookings by its ID
// @Summary      Get bookings
// @Description  Get a bookings by its ID
// @Tags         bookingss
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Bookings ID"
// @Success      200  {object}  domain.Bookings
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Bookings not found"
// @Router       /bookingss/{id} [get]
func (h *BookingsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.ParseUint(vars["id"], 10, 32)
        if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        entity, err := h.service.GetBookingsByID(uint(id))
        if err != nil {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// List retrieves all bookingss
// @Summary      List bookingss
// @Description  Get a list of all bookingss
// @Tags         bookingss
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Bookings
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /bookingss [get]
func (h *BookingsHandler) List(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		filter := domain.SearchFilter{
			Query: query,
			Limit: limit,
			Offset: offset,
		}

        entities, err := h.service.ListBookingss(filter)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entities)
}
