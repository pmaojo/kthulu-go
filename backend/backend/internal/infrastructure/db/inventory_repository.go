// @kthulu:module:inventory
package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// inventoryRepository implements the InventoryRepository interface using GORM
type inventoryRepository struct {
	db *gorm.DB
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(db *gorm.DB) repository.InventoryRepository {
	return &inventoryRepository{db: db}
}

// Warehouse operations
func (r *inventoryRepository) CreateWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	return r.db.WithContext(ctx).Create(warehouse).Error
}

func (r *inventoryRepository) GetWarehouse(ctx context.Context, id uint) (*domain.Warehouse, error) {
	var warehouse domain.Warehouse
	err := r.db.WithContext(ctx).
		Preload("InventoryItems").
		First(&warehouse, id).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

func (r *inventoryRepository) ListWarehouses(ctx context.Context, page, pageSize int, search string) ([]domain.Warehouse, int, error) {
	var warehouses []domain.Warehouse
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Warehouse{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("name ASC").
		Find(&warehouses).Error

	return warehouses, int(total), err
}

func (r *inventoryRepository) UpdateWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	return r.db.WithContext(ctx).Save(warehouse).Error
}

func (r *inventoryRepository) DeleteWarehouse(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Warehouse{}, id).Error
}

// Inventory item operations
func (r *inventoryRepository) CreateInventoryItem(ctx context.Context, item *domain.InventoryItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *inventoryRepository) GetInventoryItem(ctx context.Context, warehouseID, productID uint) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	err := r.db.WithContext(ctx).
		Preload("Warehouse").
		Preload("Product").
		Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *inventoryRepository) GetInventoryItemByID(ctx context.Context, id uint) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	err := r.db.WithContext(ctx).
		Preload("Warehouse").
		Preload("Product").
		First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *inventoryRepository) ListInventoryItems(ctx context.Context, warehouseID *uint, page, pageSize int) ([]domain.InventoryItem, int, error) {
	var items []domain.InventoryItem
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.InventoryItem{}).
		Preload("Warehouse").
		Preload("Product")

	if warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("sku ASC").
		Find(&items).Error

	return items, int(total), err
}

func (r *inventoryRepository) UpdateInventoryItem(ctx context.Context, item *domain.InventoryItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *inventoryRepository) DeleteInventoryItem(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.InventoryItem{}, id).Error
}

// Stock movement operations
func (r *inventoryRepository) CreateStockMovement(ctx context.Context, movement *domain.StockMovement) error {
	return r.db.WithContext(ctx).Create(movement).Error
}

func (r *inventoryRepository) GetStockMovements(ctx context.Context, inventoryItemID uint, page, pageSize int) ([]domain.StockMovement, int, error) {
	var movements []domain.StockMovement
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.StockMovement{}).
		Preload("User").
		Where("inventory_item_id = ?", inventoryItemID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&movements).Error

	return movements, int(total), err
}

func (r *inventoryRepository) GetStockMovementsByWarehouse(ctx context.Context, warehouseID uint, page, pageSize int) ([]domain.StockMovement, int, error) {
	var movements []domain.StockMovement
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.StockMovement{}).
		Preload("User").
		Preload("InventoryItem").
		Joins("JOIN inventory_items ON stock_movements.inventory_item_id = inventory_items.id").
		Where("inventory_items.warehouse_id = ?", warehouseID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("stock_movements.created_at DESC").
		Find(&movements).Error

	return movements, int(total), err
}

// Stock adjustment operations
func (r *inventoryRepository) CreateStockAdjustment(ctx context.Context, adjustment *domain.StockAdjustment) error {
	return r.db.WithContext(ctx).Create(adjustment).Error
}

func (r *inventoryRepository) GetStockAdjustments(ctx context.Context, inventoryItemID uint, page, pageSize int) ([]domain.StockAdjustment, int, error) {
	var adjustments []domain.StockAdjustment
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.StockAdjustment{}).
		Preload("User").
		Where("inventory_item_id = ?", inventoryItemID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&adjustments).Error

	return adjustments, int(total), err
}

// Reporting and analytics
func (r *inventoryRepository) GetLowStockItems(ctx context.Context, warehouseID *uint) ([]domain.InventoryItem, error) {
	query := r.db.WithContext(ctx).
		Preload("Warehouse").
		Preload("Product").
		Where("quantity <= reorder_point")

	if warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	var items []domain.InventoryItem
	err := query.Find(&items).Error
	return items, err
}

func (r *inventoryRepository) GetOutOfStockItems(ctx context.Context, warehouseID *uint) ([]domain.InventoryItem, error) {
	query := r.db.WithContext(ctx).
		Preload("Warehouse").
		Preload("Product").
		Where("quantity <= 0")

	if warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	var items []domain.InventoryItem
	err := query.Find(&items).Error
	return items, err
}

func (r *inventoryRepository) GetInventoryValue(ctx context.Context, warehouseID *uint) (float64, error) {
	query := r.db.WithContext(ctx).Model(&domain.InventoryItem{}).
		Select("COALESCE(SUM(quantity * unit_cost), 0) as total_value")

	if warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	var totalValue float64
	err := query.Scan(&totalValue).Error
	return totalValue, err
}

func (r *inventoryRepository) GetStockLevels(ctx context.Context, warehouseID *uint, productIDs []uint) ([]domain.InventoryItem, error) {
	query := r.db.WithContext(ctx).
		Preload("Warehouse").
		Preload("Product")

	if warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	if len(productIDs) > 0 {
		query = query.Where("product_id IN ?", productIDs)
	}

	var items []domain.InventoryItem
	err := query.Find(&items).Error
	return items, err
}
