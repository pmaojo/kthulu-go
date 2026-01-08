// @kthulu:handler:ProductHandler
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"demo-app/internal/modules/ProductHandler/domain"
	
)

type ProductHandlerHandler struct {
	service domain.ProductHandlerService
}

func NewProductHandlerHandler(service domain.ProductHandlerService) *ProductHandlerHandler {
	return &ProductHandlerHandler{service: service}
}

// RegisterRoutes registers all routes for ProductHandler
func (h *ProductHandlerHandler) RegisterRoutes(router *mux.Router) {
	sub := router.PathPrefix("product-handler").Subrouter()
	

	sub.HandleFunc("/product-handler", h.List).Methods("GET")
	sub.HandleFunc("/product-handler", h.Create).Methods("POST")
	sub.HandleFunc("/product-handler/{id}", h.GetByID).Methods("GET")
	sub.HandleFunc("/product-handler/{id}", h.Update).Methods("PUT")
	sub.HandleFunc("/product-handler/{id}", h.Delete).Methods("DELETE")
}

// Create handles the creation of a new ProductHandler
// @Summary      Create a new ProductHandler
// @Description  Creates a new ProductHandler with the provided data
// @Tags         ProductHandlers
// @Accept       json
// @Produce      json
// @Param        input body domain.ProductHandler true "ProductHandler object"
// @Success      200  {object}  domain.ProductHandler
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /product-handler [post]
func (h *ProductHandlerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var entity domain.ProductHandler
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := h.service.CreateProductHandler(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a ProductHandler by its ID
// @Summary      Get ProductHandler
// @Description  Get a ProductHandler by its ID
// @Tags         ProductHandlers
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ProductHandler ID"
// @Success      200  {object}  domain.ProductHandler
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "ProductHandler not found"
// @Router       /product-handler/{id} [get]
func (h *ProductHandlerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	entity, err := h.service.GetProductHandlerByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Update handles the update of a ProductHandler
func (h *ProductHandlerHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var entity domain.ProductHandler
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	entity.ID = uint(id)
	if err := h.service.UpdateProductHandler(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Delete handles the deletion of a ProductHandler
func (h *ProductHandlerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	if err := h.service.DeleteProductHandler(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// List retrieves all ProductHandlers
func (h *ProductHandlerHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := domain.SearchFilter{
		Query: query,
		Limit: limit,
		Offset: offset,
	}

	entities, err := h.service.ListProductHandlers(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
