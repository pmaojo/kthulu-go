// @kthulu:module:inventory
package usecase

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"backend/internal/domain"
	"backend/internal/repository"
)

// InventoryUseCase handles inventory management business logic
type InventoryUseCase struct {
	inventoryRepo repository.InventoryRepository
	productRepo   repository.ProductRepository
	userRepo      repository.UserRepository
	logger        *zap.Logger
}

// NewInventoryUseCase creates a new inventory use case
func NewInventoryUseCase(
	inventoryRepo repository.InventoryRepository,
	productRepo repository.ProductRepository,
	userRepo repository.UserRepository,
	logger *zap.Logger,
) *InventoryUseCase {
	return &InventoryUseCase{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		userRepo:      userRepo,
		logger:        logger,
	}
}

// CreateWarehouse creates a new warehouse
func (uc *InventoryUseCase) CreateWarehouse(ctx context.Context, req CreateWarehouseRequest) (*domain.Warehouse, error) {
	warehouse := &domain.Warehouse{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Address:     req.Address,
		IsActive:    true,
	}

	if err := uc.inventoryRepo.CreateWarehouse(ctx, warehouse); err != nil {
		uc.logger.Error("Failed to create warehouse", zap.Error(err))
		return nil, err
	}

	uc.logger.Info("Warehouse created", zap.String("code", warehouse.Code))
	return warehouse, nil
}

// GetWarehouse retrieves a warehouse by ID
func (uc *InventoryUseCase) GetWarehouse(ctx context.Context, id uint) (*domain.Warehouse, error) {
	warehouse, err := uc.inventoryRepo.GetWarehouse(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get warehouse", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return warehouse, nil
}

// ListWarehouses retrieves all warehouses with pagination
func (uc *InventoryUseCase) ListWarehouses(ctx context.Context, req ListWarehousesRequest) (*PaginatedResponse[domain.Warehouse], error) {
	warehouses, total, err := uc.inventoryRepo.ListWarehouses(ctx, req.Page, req.PageSize, req.Search)
	if err != nil {
		uc.logger.Error("Failed to list warehouses", zap.Error(err))
		return nil, err
	}

	return &PaginatedResponse[domain.Warehouse]{
		Data:       warehouses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (total + req.PageSize - 1) / req.PageSize,
	}, nil
}

// UpdateWarehouse updates an existing warehouse
func (uc *InventoryUseCase) UpdateWarehouse(ctx context.Context, id uint, req UpdateWarehouseRequest) (*domain.Warehouse, error) {
	warehouse, err := uc.inventoryRepo.GetWarehouse(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		warehouse.Name = req.Name
	}
	if req.Description != "" {
		warehouse.Description = req.Description
	}
	if req.Address != "" {
		warehouse.Address = req.Address
	}
	if req.IsActive != nil {
		warehouse.IsActive = *req.IsActive
	}

	if err := uc.inventoryRepo.UpdateWarehouse(ctx, warehouse); err != nil {
		uc.logger.Error("Failed to update warehouse", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}

	uc.logger.Info("Warehouse updated", zap.Uint("id", id))
	return warehouse, nil
}

// GetInventoryItem retrieves inventory item by warehouse and product
func (uc *InventoryUseCase) GetInventoryItem(ctx context.Context, warehouseID, productID uint) (*domain.InventoryItem, error) {
	item, err := uc.inventoryRepo.GetInventoryItem(ctx, warehouseID, productID)
	if err != nil {
		uc.logger.Error("Failed to get inventory item",
			zap.Uint("warehouseId", warehouseID),
			zap.Uint("productId", productID),
			zap.Error(err))
		return nil, err
	}
	return item, nil
}

// ListInventoryItems retrieves inventory items with optional warehouse filter and pagination
func (uc *InventoryUseCase) ListInventoryItems(ctx context.Context, req ListInventoryItemsRequest) (*PaginatedResponse[domain.InventoryItem], error) {
	items, total, err := uc.inventoryRepo.ListInventoryItems(ctx, req.WarehouseID, req.Page, req.PageSize)
	if err != nil {
		uc.logger.Error("Failed to list inventory items", zap.Error(err))
		return nil, err
	}

	return &PaginatedResponse[domain.InventoryItem]{
		Data:       items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (total + req.PageSize - 1) / req.PageSize,
	}, nil
}

// UpdateStock updates inventory levels for a product in a warehouse
func (uc *InventoryUseCase) UpdateStock(ctx context.Context, req UpdateStockRequest) (*domain.InventoryItem, error) {
	organizationID := req.OrganizationID
	if organizationID == 0 {
		var err error
		organizationID, err = getOrganizationIDFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("organization ID is required: %w", err)
		}
	}

	// Get or create inventory item
	item, err := uc.inventoryRepo.GetInventoryItem(ctx, req.WarehouseID, req.ProductID)
	if err != nil {
		// Create new inventory item if it doesn't exist
		product, err := uc.productRepo.GetByID(ctx, organizationID, req.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product not found: %w", err)
		}

		item = &domain.InventoryItem{
			WarehouseID:      req.WarehouseID,
			ProductID:        req.ProductID,
			SKU:              product.SKU,
			Quantity:         0,
			ReservedQuantity: 0,
			MinimumStock:     req.MinimumStock,
			ReorderPoint:     req.ReorderPoint,
			UnitCost:         req.UnitCost,
		}

		if err := uc.inventoryRepo.CreateInventoryItem(ctx, item); err != nil {
			return nil, err
		}
	}

	previousQuantity := item.Quantity

	// Update quantities based on movement type
	switch req.MovementType {
	case domain.StockMovementTypeReceive:
		item.Quantity += req.Quantity
	case domain.StockMovementTypeAdjustment:
		if req.AdjustmentType == "set" {
			item.Quantity = req.Quantity
		} else if req.AdjustmentType == "increase" {
			item.Quantity += req.Quantity
		} else if req.AdjustmentType == "decrease" {
			item.Quantity -= req.Quantity
		}
	case domain.StockMovementTypeSale:
		item.Quantity -= req.Quantity
	case domain.StockMovementTypeReturn:
		item.Quantity += req.Quantity
	case domain.StockMovementTypeReserve:
		item.ReservedQuantity += req.Quantity
	case domain.StockMovementTypeRelease:
		item.ReservedQuantity -= req.Quantity
	}

	// Ensure quantities don't go negative
	if item.Quantity < 0 {
		item.Quantity = 0
	}
	if item.ReservedQuantity < 0 {
		item.ReservedQuantity = 0
	}

	// Update unit cost if provided
	if req.UnitCost > 0 {
		item.UnitCost = req.UnitCost
	}

	now := time.Now()
	item.LastStockDate = &now

	// Update inventory item
	if err := uc.inventoryRepo.UpdateInventoryItem(ctx, item); err != nil {
		return nil, err
	}

	// Create stock movement record
	movement := &domain.StockMovement{
		InventoryItemID:  item.ID,
		Type:             req.MovementType,
		Quantity:         req.Quantity,
		PreviousQuantity: previousQuantity,
		NewQuantity:      item.Quantity,
		UnitCost:         req.UnitCost,
		TotalCost:        float64(req.Quantity) * req.UnitCost,
		Reference:        req.Reference,
		Notes:            req.Notes,
		UserID:           req.UserID,
	}

	if err := uc.inventoryRepo.CreateStockMovement(ctx, movement); err != nil {
		uc.logger.Error("Failed to create stock movement", zap.Error(err))
		// Don't fail the whole operation if movement logging fails
	}

	uc.logger.Info("Stock updated",
		zap.Uint("warehouseId", req.WarehouseID),
		zap.Uint("productId", req.ProductID),
		zap.String("movementType", string(req.MovementType)),
		zap.Int("quantity", req.Quantity))

	return item, nil
}

// GetLowStockItems retrieves items that are below their reorder point
func (uc *InventoryUseCase) GetLowStockItems(ctx context.Context, warehouseID *uint) ([]domain.InventoryItem, error) {
	items, err := uc.inventoryRepo.GetLowStockItems(ctx, warehouseID)
	if err != nil {
		uc.logger.Error("Failed to get low stock items", zap.Error(err))
		return nil, err
	}
	return items, nil
}

// GetStockMovements retrieves stock movement history
func (uc *InventoryUseCase) GetStockMovements(ctx context.Context, req GetStockMovementsRequest) (*PaginatedResponse[domain.StockMovement], error) {
	movements, total, err := uc.inventoryRepo.GetStockMovements(ctx, req.InventoryItemID, req.Page, req.PageSize)
	if err != nil {
		uc.logger.Error("Failed to get stock movements", zap.Error(err))
		return nil, err
	}

	return &PaginatedResponse[domain.StockMovement]{
		Data:       movements,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (total + req.PageSize - 1) / req.PageSize,
	}, nil
}

// Helper function to extract organization ID from context
func getOrganizationIDFromContext(ctx context.Context) (uint, error) {
	if orgID, ok := ctx.Value("organizationID").(uint); ok && orgID != 0 {
		return orgID, nil
	}
	return 0, fmt.Errorf("organization ID not found in context")
}

// Request/Response DTOs
type CreateWarehouseRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	Code        string `json:"code" validate:"required,max=50"`
	Description string `json:"description" validate:"max=500"`
	Address     string `json:"address" validate:"max=500"`
}

type UpdateWarehouseRequest struct {
	Name        string `json:"name" validate:"max=255"`
	Description string `json:"description" validate:"max=500"`
	Address     string `json:"address" validate:"max=500"`
	IsActive    *bool  `json:"isActive"`
}

type ListWarehousesRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"pageSize" validate:"min=1,max=100"`
	Search   string `json:"search"`
}

type ListInventoryItemsRequest struct {
	WarehouseID *uint `json:"warehouseId"`
	Page        int   `json:"page" validate:"min=1"`
	PageSize    int   `json:"pageSize" validate:"min=1,max=100"`
}

type UpdateStockRequest struct {
	OrganizationID uint                     `json:"-"`
	WarehouseID    uint                     `json:"warehouseId" validate:"required"`
	ProductID      uint                     `json:"productId" validate:"required"`
	MovementType   domain.StockMovementType `json:"movementType" validate:"required"`
	Quantity       int                      `json:"quantity" validate:"required"`
	AdjustmentType string                   `json:"adjustmentType"` // for adjustments: set, increase, decrease
	UnitCost       float64                  `json:"unitCost" validate:"min=0"`
	MinimumStock   int                      `json:"minimumStock" validate:"min=0"`
	ReorderPoint   int                      `json:"reorderPoint" validate:"min=0"`
	Reference      string                   `json:"reference" validate:"max=255"`
	Notes          string                   `json:"notes" validate:"max=500"`
	UserID         *uint                    `json:"userId"`
}

type GetStockMovementsRequest struct {
	InventoryItemID uint `json:"inventoryItemId" validate:"required"`
	Page            int  `json:"page" validate:"min=1"`
	PageSize        int  `json:"pageSize" validate:"min=1,max=100"`
}
