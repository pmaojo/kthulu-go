// @kthulu:handler:customers
package handlers

import (
        "encoding/json"
        "net/http"
        "strconv"

        "github.com/gorilla/mux"
        "github.com/example/workshop-api/internal/adapters/http/modules/customers/domain"
)

type CustomersHandler struct {
        service domain.CustomersService
}

func NewCustomersHandler(service domain.CustomersService) *CustomersHandler {
        return &CustomersHandler{service: service}
}

func (h *CustomersHandler) RegisterRoutes(r *mux.Router) {
        r.HandleFunc("/customerss", h.List).Methods("GET")
        r.HandleFunc("/customerss", h.Create).Methods("POST")
        r.HandleFunc("/customerss/{id}", h.GetByID).Methods("GET")
}

// Create handles the creation of a new customers
// @Summary      Create a new customers
// @Description  Creates a new customers with the provided data
// @Tags         customerss
// @Accept       json
// @Produce      json
// @Param        input body domain.Customers true "Customers object"
// @Success      200  {object}  domain.Customers
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /customerss [post]
func (h *CustomersHandler) Create(w http.ResponseWriter, r *http.Request) {
        var entity domain.Customers
        if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }

        if err := h.service.CreateCustomers(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a customers by its ID
// @Summary      Get customers
// @Description  Get a customers by its ID
// @Tags         customerss
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Customers ID"
// @Success      200  {object}  domain.Customers
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Customers not found"
// @Router       /customerss/{id} [get]
func (h *CustomersHandler) GetByID(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.ParseUint(vars["id"], 10, 32)
        if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        entity, err := h.service.GetCustomersByID(uint(id))
        if err != nil {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// List retrieves all customerss
// @Summary      List customerss
// @Description  Get a list of all customerss
// @Tags         customerss
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Customers
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /customerss [get]
func (h *CustomersHandler) List(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		filter := domain.SearchFilter{
			Query: query,
			Limit: limit,
			Offset: offset,
		}

        entities, err := h.service.ListCustomerss(filter)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entities)
}
