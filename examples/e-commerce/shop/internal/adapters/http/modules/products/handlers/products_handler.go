// @kthulu:handler:products
package handlers

import (
        "encoding/json"
        "net/http"
        "strconv"

        "github.com/gorilla/mux"
        "shop/internal/adapters/http/modules/products/domain"
)

type ProductsHandler struct {
        service domain.ProductsService
}

func NewProductsHandler(service domain.ProductsService) *ProductsHandler {
        return &ProductsHandler{service: service}
}

// Create handles the creation of a new products
// @Summary      Create a new products
// @Description  Creates a new products with the provided data
// @Tags         productss
// @Accept       json
// @Produce      json
// @Param        input body domain.Products true "Products object"
// @Success      200  {object}  domain.Products
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /productss [post]
func (h *ProductsHandler) Create(w http.ResponseWriter, r *http.Request) {
        var entity domain.Products
        if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }

        if err := h.service.CreateProducts(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a products by its ID
// @Summary      Get products
// @Description  Get a products by its ID
// @Tags         productss
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Products ID"
// @Success      200  {object}  domain.Products
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "Products not found"
// @Router       /productss/{id} [get]
func (h *ProductsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.ParseUint(vars["id"], 10, 32)
        if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        entity, err := h.service.GetProductsByID(uint(id))
        if err != nil {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// List retrieves all productss
// @Summary      List productss
// @Description  Get a list of all productss
// @Tags         productss
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Products
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /productss [get]
func (h *ProductsHandler) List(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		filter := domain.SearchFilter{
			Query: query,
			Limit: limit,
			Offset: offset,
		}

        entities, err := h.service.ListProductss(filter)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entities)
}
