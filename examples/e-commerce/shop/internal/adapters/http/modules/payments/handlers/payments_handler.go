// @kthulu:handler:payments
package handlers

import (
        "encoding/json"
        "net/http"
        "strconv"

        "github.com/gorilla/mux"
        "shop/internal/adapters/http/modules/payments/domain"
)

type PaymentsHandler struct {
        service domain.PaymentsService
}

func NewPaymentsHandler(service domain.PaymentsService) *PaymentsHandler {
        return &PaymentsHandler{service: service}
}

// Create handles the creation of a new payments
// @Summary      Create a new payments
// @Description  Creates a new payments with the provided data
// @Tags         paymentss
// @Accept       json
// @Produce      json
// @Param        input body domain.Payments true "Payments object"
// @Success      200  {object}  domain.Payments
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /paymentss [post]
func (h *PaymentsHandler) Create(w http.ResponseWriter, r *http.Request) {
        var entity domain.Payments
        if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }

        if err := h.service.CreatePayments(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a payments by its ID
// @Summary      Get payments
// @Description  Get a payments by its ID
// @Tags         paymentss
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Payments ID"
// @Success      200  {object}  domain.Payments
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Payments not found"
// @Router       /paymentss/{id} [get]
func (h *PaymentsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.ParseUint(vars["id"], 10, 32)
        if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        entity, err := h.service.GetPaymentsByID(uint(id))
        if err != nil {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// List retrieves all paymentss
// @Summary      List paymentss
// @Description  Get a list of all paymentss
// @Tags         paymentss
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Payments
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /paymentss [get]
func (h *PaymentsHandler) List(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		filter := domain.SearchFilter{
			Query: query,
			Limit: limit,
			Offset: offset,
		}

        entities, err := h.service.ListPaymentss(filter)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entities)
}
