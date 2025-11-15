// @kthulu:module:inventory
package repository

import (
	"context"

	"github.com/kthulu/kthulu-go/backend/internal/domain"
)

// InventoryRepository defines the interface for inventory data access
type InventoryRepository interface {
	// Warehouse operations
	CreateWarehouse(ctx context.Context, warehouse *domain.Warehouse) error
	GetWarehouse(ctx context.Context, id uint) (*domain.Warehouse, error)
	ListWarehouses(ctx context.Context, page, pageSize int, search string) ([]domain.Warehouse, int, error)
	UpdateWarehouse(ctx context.Context, warehouse *domain.Warehouse) error
	DeleteWarehouse(ctx context.Context, id uint) error

	// Inventory item operations
	CreateInventoryItem(ctx context.Context, item *domain.InventoryItem) error
	GetInventoryItem(ctx context.Context, warehouseID, productID uint) (*domain.InventoryItem, error)
	GetInventoryItemByID(ctx context.Context, id uint) (*domain.InventoryItem, error)
	ListInventoryItems(ctx context.Context, warehouseID *uint, page, pageSize int) ([]domain.InventoryItem, int, error)
	UpdateInventoryItem(ctx context.Context, item *domain.InventoryItem) error
	DeleteInventoryItem(ctx context.Context, id uint) error

	// Stock movement operations
	CreateStockMovement(ctx context.Context, movement *domain.StockMovement) error
	GetStockMovements(ctx context.Context, inventoryItemID uint, page, pageSize int) ([]domain.StockMovement, int, error)
	GetStockMovementsByWarehouse(ctx context.Context, warehouseID uint, page, pageSize int) ([]domain.StockMovement, int, error)

	// Stock adjustment operations
	CreateStockAdjustment(ctx context.Context, adjustment *domain.StockAdjustment) error
	GetStockAdjustments(ctx context.Context, inventoryItemID uint, page, pageSize int) ([]domain.StockAdjustment, int, error)

	// Reporting and analytics
	GetLowStockItems(ctx context.Context, warehouseID *uint) ([]domain.InventoryItem, error)
	GetOutOfStockItems(ctx context.Context, warehouseID *uint) ([]domain.InventoryItem, error)
	GetInventoryValue(ctx context.Context, warehouseID *uint) (float64, error)
	GetStockLevels(ctx context.Context, warehouseID *uint, productIDs []uint) ([]domain.InventoryItem, error)
}
