// @kthulu:handler:orders
package handlers

import (
        "encoding/json"
        "net/http"
        "strconv"

        "github.com/gorilla/mux"
        "shop/internal/adapters/http/modules/orders/domain"
)

type OrdersHandler struct {
        service domain.OrdersService
}

func NewOrdersHandler(service domain.OrdersService) *OrdersHandler {
        return &OrdersHandler{service: service}
}

// Create handles the creation of a new orders
// @Summary      Create a new orders
// @Description  Creates a new orders with the provided data
// @Tags         orderss
// @Accept       json
// @Produce      json
// @Param        input body domain.Orders true "Orders object"
// @Success      200  {object}  domain.Orders
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /orderss [post]
func (h *OrdersHandler) Create(w http.ResponseWriter, r *http.Request) {
        var entity domain.Orders
        if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }

        if err := h.service.CreateOrders(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a orders by its ID
// @Summary      Get orders
// @Description  Get a orders by its ID
// @Tags         orderss
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Orders ID"
// @Success      200  {object}  domain.Orders
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Orders not found"
// @Router       /orderss/{id} [get]
func (h *OrdersHandler) GetByID(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.ParseUint(vars["id"], 10, 32)
        if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        entity, err := h.service.GetOrdersByID(uint(id))
        if err != nil {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// List retrieves all orderss
// @Summary      List orderss
// @Description  Get a list of all orderss
// @Tags         orderss
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Orders
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /orderss [get]
func (h *OrdersHandler) List(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		filter := domain.SearchFilter{
			Query: query,
			Limit: limit,
			Offset: offset,
		}

        entities, err := h.service.ListOrderss(filter)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entities)
}
