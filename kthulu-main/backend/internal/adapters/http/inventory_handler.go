// @kthulu:module:inventory
package adapterhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"backend/internal/domain"
	"backend/internal/usecase"
)

// ensure domain types are recognized by documentation generators
var _ = domain.Warehouse{}

// InventoryHandler handles inventory-related HTTP requests
type InventoryHandler struct {
	inventoryUC *usecase.InventoryUseCase
	logger      *zap.Logger
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(inventoryUC *usecase.InventoryUseCase, logger *zap.Logger) *InventoryHandler {
	return &InventoryHandler{
		inventoryUC: inventoryUC,
		logger:      logger,
	}
}

// RegisterRoutes registers inventory routes
func (h *InventoryHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/inventory", func(r chi.Router) {
		// Warehouse routes
		r.Route("/warehouses", func(r chi.Router) {
			r.Post("/", h.CreateWarehouse)
			r.Get("/", h.ListWarehouses)
			r.Get("/{id}", h.GetWarehouse)
			r.Put("/{id}", h.UpdateWarehouse)
		})

		// Inventory item routes
		r.Route("/items", func(r chi.Router) {
			r.Get("/", h.ListInventoryItems)
			r.Get("/{warehouseId}/{productId}", h.GetInventoryItem)
			r.Put("/stock", h.UpdateStock)
		})

		// Stock movement routes
		r.Route("/movements", func(r chi.Router) {
			r.Get("/{inventoryItemId}", h.GetStockMovements)
		})

		// Reporting routes
		r.Get("/low-stock", h.GetLowStockItems)
	})

	h.logger.Info("Inventory routes registered")
}

// CreateWarehouse godoc
// @Summary Create a new warehouse
// @Description Creates a new warehouse for inventory management
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body usecase.CreateWarehouseRequest true "Warehouse details"
// @Success 201 {object} domain.Warehouse "Warehouse created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/inventory/warehouses [post]
func (h *InventoryHandler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode create warehouse request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	warehouse, err := h.inventoryUC.CreateWarehouse(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create warehouse", zap.Error(err))
		http.Error(w, "Failed to create warehouse", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(warehouse)
}

// GetWarehouse godoc
// @Summary Get warehouse by ID
// @Description Retrieves a warehouse by its ID
// @Tags Inventory
// @Produce json
// @Param id path int true "Warehouse ID"
// @Success 200 {object} domain.Warehouse "Warehouse details"
// @Failure 404 {object} map[string]string "Warehouse not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/inventory/warehouses/{id} [get]
func (h *InventoryHandler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	warehouse, err := h.inventoryUC.GetWarehouse(r.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get warehouse", zap.Uint64("id", id), zap.Error(err))
		http.Error(w, "Warehouse not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(warehouse)
}

// ListWarehouses godoc
// @Summary List warehouses
// @Description Retrieves a paginated list of warehouses
// @Tags Inventory
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param search query string false "Search term"
// @Success 200 {object} usecase.PaginatedResponse[domain.Warehouse] "List of warehouses"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/inventory/warehouses [get]
func (h *InventoryHandler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	search := r.URL.Query().Get("search")

	req := usecase.ListWarehousesRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	response, err := h.inventoryUC.ListWarehouses(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list warehouses", zap.Error(err))
		http.Error(w, "Failed to list warehouses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateWarehouse godoc
// @Summary Update warehouse
// @Description Updates an existing warehouse
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path int true "Warehouse ID"
// @Param request body usecase.UpdateWarehouseRequest true "Updated warehouse details"
// @Success 200 {object} domain.Warehouse "Updated warehouse"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Warehouse not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/inventory/warehouses/{id} [put]
func (h *InventoryHandler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	var req usecase.UpdateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode update warehouse request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	warehouse, err := h.inventoryUC.UpdateWarehouse(r.Context(), uint(id), req)
	if err != nil {
		h.logger.Error("Failed to update warehouse", zap.Uint64("id", id), zap.Error(err))
		http.Error(w, "Failed to update warehouse", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(warehouse)
}

// GetInventoryItem godoc
// @Summary Get inventory item
// @Description Retrieves inventory item by warehouse and product ID
// @Tags Inventory
// @Produce json
// @Param warehouseId path int true "Warehouse ID"
// @Param productId path int true "Product ID"
// @Success 200 {object} domain.InventoryItem "Inventory item details"
// @Failure 404 {object} map[string]string "Inventory item not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/inventory/items/{warehouseId}/{productId} [get]
func (h *InventoryHandler) GetInventoryItem(w http.ResponseWriter, r *http.Request) {
	warehouseIDStr := chi.URLParam(r, "warehouseId")
	productIDStr := chi.URLParam(r, "productId")

	warehouseID, err := strconv.ParseUint(warehouseIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	item, err := h.inventoryUC.GetInventoryItem(r.Context(), uint(warehouseID), uint(productID))
	if err != nil {
		h.logger.Error("Failed to get inventory item",
			zap.Uint64("warehouseId", warehouseID),
			zap.Uint64("productId", productID),
			zap.Error(err))
		http.Error(w, "Inventory item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// ListInventoryItems godoc
// @Summary List inventory items
// @Description Retrieves a paginated list of inventory items
// @Tags Inventory
// @Produce json
// @Param warehouseId query int false "Filter by warehouse ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Success 200 {object} usecase.PaginatedResponse[domain.InventoryItem] "List of inventory items"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/inventory/items [get]
func (h *InventoryHandler) ListInventoryItems(w http.ResponseWriter, r *http.Request) {

	var warehouseID *uint
	if warehouseIDStr := r.URL.Query().Get("warehouseId"); warehouseIDStr != "" {
		if id, err := strconv.ParseUint(warehouseIDStr, 10, 32); err == nil {
			wid := uint(id)
			warehouseID = &wid
		}
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	req := usecase.ListInventoryItemsRequest{
		WarehouseID: warehouseID,
		Page:        page,
		PageSize:    pageSize,
	}

	response, err := h.inventoryUC.ListInventoryItems(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list inventory items", zap.Error(err))
		http.Error(w, "Failed to list inventory items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateStock godoc
// @Summary Update stock levels
// @Description Updates inventory stock levels for a product in a warehouse
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body usecase.UpdateStockRequest true "Stock update details"
// @Success 200 {object} domain.InventoryItem "Updated inventory item"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/inventory/items/stock [put]
func (h *InventoryHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	var req usecase.UpdateStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode update stock request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.OrganizationID = h.getOrganizationID(r)

	item, err := h.inventoryUC.UpdateStock(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to update stock", zap.Error(err))
		http.Error(w, "Failed to update stock", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// GetStockMovements godoc
// @Summary Get stock movements
// @Description Retrieves stock movement history for an inventory item
// @Tags Inventory
// @Produce json
// @Param inventoryItemId path int true "Inventory Item ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Success 200 {object} usecase.PaginatedResponse[domain.StockMovement] "Stock movement history"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/inventory/movements/{inventoryItemId} [get]
func (h *InventoryHandler) GetStockMovements(w http.ResponseWriter, r *http.Request) {
	inventoryItemIDStr := chi.URLParam(r, "inventoryItemId")
	inventoryItemID, err := strconv.ParseUint(inventoryItemIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid inventory item ID", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	req := usecase.GetStockMovementsRequest{
		InventoryItemID: uint(inventoryItemID),
		Page:            page,
		PageSize:        pageSize,
	}

	response, err := h.inventoryUC.GetStockMovements(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get stock movements", zap.Error(err))
		http.Error(w, "Failed to get stock movements", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLowStockItems godoc
// @Summary Get low stock items
// @Description Retrieves items that are below their reorder point
// @Tags Inventory
// @Produce json
// @Param warehouseId query int false "Filter by warehouse ID"
// @Success 200 {array} domain.InventoryItem "Low stock items"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/inventory/low-stock [get]
func (h *InventoryHandler) GetLowStockItems(w http.ResponseWriter, r *http.Request) {
	var warehouseID *uint
	if warehouseIDStr := r.URL.Query().Get("warehouseId"); warehouseIDStr != "" {
		if id, err := strconv.ParseUint(warehouseIDStr, 10, 32); err == nil {
			wid := uint(id)
			warehouseID = &wid
		}
	}

	items, err := h.inventoryUC.GetLowStockItems(r.Context(), warehouseID)
	if err != nil {
		h.logger.Error("Failed to get low stock items", zap.Error(err))
		http.Error(w, "Failed to get low stock items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *InventoryHandler) getOrganizationID(r *http.Request) uint {
	if orgIDStr := r.Header.Get("X-Organization-ID"); orgIDStr != "" {
		if orgID, err := strconv.ParseUint(orgIDStr, 10, 32); err == nil {
			return uint(orgID)
		}
	}
	return 0
}
