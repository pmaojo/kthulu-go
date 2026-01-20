// @kthulu:core:inventory
package core

import (
	"time"
	
)

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Inventory represents a inventory entity
type Inventory struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Product_id int `json:"product_id,string" gorm:"product_id"`
	Quantity int `json:"quantity,string" gorm:"quantity"`
	Warehouse string `json:"warehouse" gorm:"warehouse"`
}

// TableName overrides the table name used by Inventory to `Inventories`
func (Inventory) TableName() string {
	return "Inventories"
}

// InventoryRepository defines the repository interface
type InventoryRepository interface {
	Create(entity *Inventory) error
	GetByID(id uint) (*Inventory, error)
	Update(entity *Inventory) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Inventory, error)
}

// InventoryService defines the service interface  
type InventoryService interface {
	CreateInventory(entity *Inventory) error
	GetInventoryByID(id uint) (*Inventory, error)
	UpdateInventory(entity *Inventory) error
	DeleteInventory(id uint) error
	ListInventories(filter SearchFilter) ([]*Inventory, error)
}
