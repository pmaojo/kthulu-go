// @kthulu:module:inventory
package domain

import (
	"time"
)

// Warehouse represents a physical or logical storage location
type Warehouse struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;size:255"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null;size:50"`
	Description string    `json:"description" gorm:"size:500"`
	Address     string    `json:"address" gorm:"size:500"`
	IsActive    bool      `json:"isActive" gorm:"default:true"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Relationships
	InventoryItems []InventoryItem `json:"inventoryItems,omitempty" gorm:"foreignKey:WarehouseID"`
}

// InventoryItem represents stock levels for a product in a warehouse
type InventoryItem struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	WarehouseID      uint       `json:"warehouseId" gorm:"not null;index"`
	ProductID        uint       `json:"productId" gorm:"not null;index"`
	SKU              string     `json:"sku" gorm:"not null;size:100;index"`
	Quantity         int        `json:"quantity" gorm:"default:0"`
	ReservedQuantity int        `json:"reservedQuantity" gorm:"default:0"`
	MinimumStock     int        `json:"minimumStock" gorm:"default:0"`
	MaximumStock     int        `json:"maximumStock" gorm:"default:0"`
	ReorderPoint     int        `json:"reorderPoint" gorm:"default:0"`
	ReorderQuantity  int        `json:"reorderQuantity" gorm:"default:0"`
	UnitCost         float64    `json:"unitCost" gorm:"type:decimal(10,2);default:0"`
	LastStockDate    *time.Time `json:"lastStockDate"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`

	// Relationships
	Warehouse      Warehouse       `json:"warehouse,omitempty" gorm:"foreignKey:WarehouseID"`
	Product        Product         `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	StockMovements []StockMovement `json:"stockMovements,omitempty" gorm:"foreignKey:InventoryItemID"`
}

// StockMovementType represents the type of stock movement
type StockMovementType string

const (
	StockMovementTypeReceive    StockMovementType = "receive"
	StockMovementTypeTransfer   StockMovementType = "transfer"
	StockMovementTypeAdjustment StockMovementType = "adjustment"
	StockMovementTypeSale       StockMovementType = "sale"
	StockMovementTypeReturn     StockMovementType = "return"
	StockMovementTypeReserve    StockMovementType = "reserve"
	StockMovementTypeRelease    StockMovementType = "release"
)

// StockMovement represents a change in inventory levels
type StockMovement struct {
	ID               uint              `json:"id" gorm:"primaryKey"`
	InventoryItemID  uint              `json:"inventoryItemId" gorm:"not null;index"`
	Type             StockMovementType `json:"type" gorm:"not null;size:20"`
	Quantity         int               `json:"quantity" gorm:"not null"`
	PreviousQuantity int               `json:"previousQuantity" gorm:"not null"`
	NewQuantity      int               `json:"newQuantity" gorm:"not null"`
	UnitCost         float64           `json:"unitCost" gorm:"type:decimal(10,2);default:0"`
	TotalCost        float64           `json:"totalCost" gorm:"type:decimal(10,2);default:0"`
	Reference        string            `json:"reference" gorm:"size:255"`
	Notes            string            `json:"notes" gorm:"size:500"`
	UserID           *uint             `json:"userId" gorm:"index"`
	CreatedAt        time.Time         `json:"createdAt"`

	// Relationships
	InventoryItem InventoryItem `json:"inventoryItem,omitempty" gorm:"foreignKey:InventoryItemID"`
	User          *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// StockAdjustment represents a manual adjustment to inventory levels
type StockAdjustment struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	InventoryItemID uint      `json:"inventoryItemId" gorm:"not null;index"`
	AdjustmentType  string    `json:"adjustmentType" gorm:"not null;size:50"` // increase, decrease, set
	Quantity        int       `json:"quantity" gorm:"not null"`
	Reason          string    `json:"reason" gorm:"not null;size:255"`
	Notes           string    `json:"notes" gorm:"size:500"`
	UserID          uint      `json:"userId" gorm:"not null;index"`
	CreatedAt       time.Time `json:"createdAt"`

	// Relationships
	InventoryItem InventoryItem `json:"inventoryItem,omitempty" gorm:"foreignKey:InventoryItemID"`
	User          User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// AvailableQuantity returns the quantity available for sale (total - reserved)
func (ii *InventoryItem) AvailableQuantity() int {
	return ii.Quantity - ii.ReservedQuantity
}

// IsLowStock checks if the item is below the reorder point
func (ii *InventoryItem) IsLowStock() bool {
	return ii.Quantity <= ii.ReorderPoint
}

// IsOutOfStock checks if the item is out of stock
func (ii *InventoryItem) IsOutOfStock() bool {
	return ii.Quantity <= 0
}

// CanFulfill checks if the item can fulfill a requested quantity
func (ii *InventoryItem) CanFulfill(requestedQuantity int) bool {
	return ii.AvailableQuantity() >= requestedQuantity
}
