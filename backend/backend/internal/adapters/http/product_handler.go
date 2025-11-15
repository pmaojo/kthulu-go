// @kthulu:module:products
package adapterhttp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/kthulu/kthulu-go/backend/internal/domain"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
	"github.com/kthulu/kthulu-go/backend/internal/usecase"
)

// ProductHandler handles HTTP requests for product operations
type ProductHandler struct {
	productUseCase *usecase.ProductUseCase
	validator      *validator.Validate
	logger         *zap.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(productUseCase *usecase.ProductUseCase, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		productUseCase: productUseCase,
		validator:      validator.New(),
		logger:         logger,
	}
}

// RegisterRoutes registers product routes
func (h *ProductHandler) RegisterRoutes(r chi.Router) {
	r.Route("/products", func(r chi.Router) {
		r.Post("/", h.CreateProduct)
		r.Get("/", h.ListProducts)
		r.Get("/stats", h.GetProductStats)
		r.Get("/{productId}", h.GetProduct)
		r.Put("/{productId}", h.UpdateProduct)
		r.Delete("/{productId}", h.DeleteProduct)
		r.Patch("/{productId}/status", h.SetProductStatus)

		// Variant routes
		r.Post("/{productId}/variants", h.CreateProductVariant)
		r.Get("/{productId}/variants", h.GetProductVariants)
		r.Get("/{productId}/variants/{variantId}", h.GetProductVariant)
		r.Put("/{productId}/variants/{variantId}", h.UpdateProductVariant)
		r.Delete("/{productId}/variants/{variantId}", h.DeleteProductVariant)

		// Price routes
		r.Post("/prices", h.CreateProductPrice)
		r.Get("/{productId}/prices", h.GetProductPrices)
		r.Get("/variants/{variantId}/prices", h.GetVariantPrices)
		r.Get("/effective-price", h.GetEffectivePrice)
		r.Put("/prices/{priceId}", h.UpdateProductPrice)
		r.Delete("/prices/{priceId}", h.DeleteProductPrice)
	})
}

// CreateProduct creates a new product
// @Summary Create a new product
// @Description Create a new product in the organization's catalog
// @Tags products
// @Accept json
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Param product body usecase.CreateProductRequest true "Product data"
// @Success 201 {object} domain.Product
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /products [post]
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	var req usecase.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.writeError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	product, err := h.productUseCase.CreateProduct(r.Context(), organizationID, req)
	if err != nil {
		switch err {
		case domain.ErrProductAlreadyExists:
			h.writeError(w, http.StatusConflict, "product with SKU already exists", err)
		default:
			h.logger.Error("Failed to create product", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to create product", err)
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, product)
}

// GetProduct retrieves a product by ID
// @Summary Get a product by ID
// @Description Retrieve a product by its ID
// @Tags products
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Param productId path string true "Product ID"
// @Param include query string false "Include related data (variants,prices)"
// @Success 200 {object} domain.Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /products/{productId} [get]
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	productID, err := h.getUintParam(r, "productId")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product ID", err)
		return
	}

	includeRelated := r.URL.Query().Get("include") != ""

	product, err := h.productUseCase.GetProduct(r.Context(), organizationID, productID, includeRelated)
	if err != nil {
		switch err {
		case domain.ErrProductNotFound:
			h.writeError(w, http.StatusNotFound, "product not found", err)
		default:
			h.logger.Error("Failed to get product", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to get product", err)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, product)
}

// UpdateProduct updates an existing product
// @Summary Update a product
// @Description Update an existing product's information
// @Tags products
// @Accept json
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Param productId path string true "Product ID"
// @Param product body usecase.UpdateProductRequest true "Updated product data"
// @Success 200 {object} domain.Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /products/{productId} [put]
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	productID, err := h.getUintParam(r, "productId")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product ID", err)
		return
	}

	var req usecase.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.writeError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	product, err := h.productUseCase.UpdateProduct(r.Context(), organizationID, productID, req)
	if err != nil {
		switch err {
		case domain.ErrProductNotFound:
			h.writeError(w, http.StatusNotFound, "product not found", err)
		default:
			h.logger.Error("Failed to update product", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to update product", err)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, product)
}

// DeleteProduct deletes a product
// @Summary Delete a product
// @Description Delete a product from the catalog
// @Tags products
// @Param organizationId header string true "Organization ID"
// @Param productId path string true "Product ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /products/{productId} [delete]
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	productID, err := h.getUintParam(r, "productId")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product ID", err)
		return
	}

	err = h.productUseCase.DeleteProduct(r.Context(), organizationID, productID)
	if err != nil {
		switch err {
		case domain.ErrProductNotFound:
			h.writeError(w, http.StatusNotFound, "product not found", err)
		default:
			h.logger.Error("Failed to delete product", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to delete product", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListProducts retrieves products with filtering and pagination
// @Summary List products
// @Description Retrieve a paginated list of products with optional filtering
// @Tags products
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Page size (default: 20, max: 100)"
// @Param category query string false "Filter by category"
// @Param brand query string false "Filter by brand"
// @Param isActive query bool false "Filter by active status"
// @Param isTrackable query bool false "Filter by trackable status"
// @Param search query string false "Search in name, SKU, description"
// @Param sortBy query string false "Sort by field (name, sku, category, brand, created_at, updated_at)"
// @Param sortOrder query string false "Sort order (asc, desc)"
// @Param includeVariants query bool false "Include product variants"
// @Param includePrices query bool false "Include product prices"
// @Success 200 {object} usecase.ProductListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /products [get]
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	filters := h.parseProductFilters(r)

	response, err := h.productUseCase.ListProducts(r.Context(), organizationID, filters)
	if err != nil {
		h.logger.Error("Failed to list products", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to list products", err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// SetProductStatus sets the active status of a product
// @Summary Set product status
// @Description Set the active status of a product
// @Tags products
// @Accept json
// @Param organizationId header string true "Organization ID"
// @Param productId path string true "Product ID"
// @Param status body object{active:bool} true "Status data"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /products/{productId}/status [patch]
func (h *ProductHandler) SetProductStatus(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	productID, err := h.getUintParam(r, "productId")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product ID", err)
		return
	}

	var req struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	err = h.productUseCase.SetProductActive(r.Context(), organizationID, productID, req.Active)
	if err != nil {
		switch err {
		case domain.ErrProductNotFound:
			h.writeError(w, http.StatusNotFound, "product not found", err)
		default:
			h.logger.Error("Failed to set product status", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to set product status", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetProductStats retrieves product statistics
// @Summary Get product statistics
// @Description Retrieve statistics for products in the organization
// @Tags products
// @Produce json
// @Param organizationId header string true "Organization ID"
// @Success 200 {object} repository.ProductStats
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /products/stats [get]
func (h *ProductHandler) GetProductStats(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	stats, err := h.productUseCase.GetProductStats(r.Context(), organizationID)
	if err != nil {
		h.logger.Error("Failed to get product stats", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to get product stats", err)
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// CreateProductVariant creates a new product variant
func (h *ProductHandler) CreateProductVariant(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	productID, err := h.getUintParam(r, "productId")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product ID", err)
		return
	}

	var req usecase.CreateVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.writeError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	variant, err := h.productUseCase.CreateProductVariant(r.Context(), organizationID, productID, req)
	if err != nil {
		switch err {
		case domain.ErrProductNotFound:
			h.writeError(w, http.StatusNotFound, "product not found", err)
		case domain.ErrVariantAlreadyExists:
			h.writeError(w, http.StatusConflict, "variant with SKU already exists", err)
		default:
			h.logger.Error("Failed to create product variant", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to create variant", err)
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, variant)
}

// CreateProductPrice creates a new product price
func (h *ProductHandler) CreateProductPrice(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	var req usecase.CreatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.writeError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	price, err := h.productUseCase.CreateProductPrice(r.Context(), organizationID, req)
	if err != nil {
		switch err {
		case domain.ErrProductNotFound:
			h.writeError(w, http.StatusNotFound, "product not found", err)
		case domain.ErrVariantNotFound:
			h.writeError(w, http.StatusNotFound, "variant not found", err)
		default:
			h.logger.Error("Failed to create product price", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to create price", err)
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, price)
}

// GetEffectivePrice gets the effective price for a product or variant
func (h *ProductHandler) GetEffectivePrice(w http.ResponseWriter, r *http.Request) {
	organizationID := h.getOrganizationID(r)
	if organizationID == 0 {
		h.writeError(w, http.StatusBadRequest, "missing organization ID", nil)
		return
	}

	var req usecase.GetEffectivePriceRequest

	// Parse query parameters
	if productIDStr := r.URL.Query().Get("productId"); productIDStr != "" {
		if productID, err := strconv.ParseUint(productIDStr, 10, 32); err == nil {
			id := uint(productID)
			req.ProductID = &id
		}
	}

	if variantIDStr := r.URL.Query().Get("variantId"); variantIDStr != "" {
		if variantID, err := strconv.ParseUint(variantIDStr, 10, 32); err == nil {
			id := uint(variantID)
			req.ProductVariantID = &id
		}
	}

	req.PriceType = domain.PriceType(r.URL.Query().Get("priceType"))
	if req.PriceType == "" {
		req.PriceType = domain.PriceTypeBase
	}

	if quantityStr := r.URL.Query().Get("quantity"); quantityStr != "" {
		if quantity, err := strconv.Atoi(quantityStr); err == nil {
			req.Quantity = quantity
		} else {
			req.Quantity = 1
		}
	} else {
		req.Quantity = 1
	}

	if atStr := r.URL.Query().Get("at"); atStr != "" {
		if at, err := time.Parse(time.RFC3339, atStr); err == nil {
			req.At = &at
		}
	}

	if err := h.validator.Struct(req); err != nil {
		h.writeError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	price, err := h.productUseCase.GetEffectivePrice(r.Context(), organizationID, req)
	if err != nil {
		switch err {
		case domain.ErrProductNotFound:
			h.writeError(w, http.StatusNotFound, "product not found", err)
		case domain.ErrVariantNotFound:
			h.writeError(w, http.StatusNotFound, "variant not found", err)
		case domain.ErrPriceNotFound:
			h.writeError(w, http.StatusNotFound, "no effective price found", err)
		default:
			h.logger.Error("Failed to get effective price", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "failed to get effective price", err)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, price)
}

// Placeholder implementations for remaining handlers
func (h *ProductHandler) GetProductVariants(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *ProductHandler) GetProductVariant(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *ProductHandler) UpdateProductVariant(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *ProductHandler) DeleteProductVariant(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *ProductHandler) GetProductPrices(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *ProductHandler) GetVariantPrices(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *ProductHandler) UpdateProductPrice(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

func (h *ProductHandler) DeleteProductPrice(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "not implemented", nil)
}

// Helper methods

func (h *ProductHandler) getOrganizationID(r *http.Request) uint {
	// This should be extracted from JWT token or header
	// For now, we'll use a header value
	if orgIDStr := r.Header.Get("X-Organization-ID"); orgIDStr != "" {
		if orgID, err := strconv.ParseUint(orgIDStr, 10, 32); err == nil {
			return uint(orgID)
		}
	}
	return 0
}

func (h *ProductHandler) getUintParam(r *http.Request, param string) (uint, error) {
	paramStr := chi.URLParam(r, param)
	if paramStr == "" {
		return 0, nil
	}

	id, err := strconv.ParseUint(paramStr, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}

func (h *ProductHandler) parseProductFilters(r *http.Request) repository.ProductFilters {
	filters := repository.DefaultProductFilters()

	if category := r.URL.Query().Get("category"); category != "" {
		filters.Category = category
	}

	if brand := r.URL.Query().Get("brand"); brand != "" {
		filters.Brand = brand
	}

	if isActiveStr := r.URL.Query().Get("isActive"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filters.IsActive = &isActive
		}
	}

	if isTrackableStr := r.URL.Query().Get("isTrackable"); isTrackableStr != "" {
		if isTrackable, err := strconv.ParseBool(isTrackableStr); err == nil {
			filters.IsTrackable = &isTrackable
		}
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = search
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}

	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			filters.PageSize = pageSize
		}
	}

	if sortBy := r.URL.Query().Get("sortBy"); sortBy != "" {
		filters.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sortOrder"); sortOrder != "" {
		filters.SortOrder = sortOrder
	}

	if includeVariantsStr := r.URL.Query().Get("includeVariants"); includeVariantsStr != "" {
		if includeVariants, err := strconv.ParseBool(includeVariantsStr); err == nil {
			filters.IncludeVariants = includeVariants
		}
	}

	if includePricesStr := r.URL.Query().Get("includePrices"); includePricesStr != "" {
		if includePrices, err := strconv.ParseBool(includePricesStr); err == nil {
			filters.IncludePrices = includePrices
		}
	}

	return filters
}

func (h *ProductHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *ProductHandler) writeError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"error":  message,
		"status": status,
	}

	if err != nil {
		response["details"] = err.Error()
	}

	json.NewEncoder(w).Encode(response)
}
