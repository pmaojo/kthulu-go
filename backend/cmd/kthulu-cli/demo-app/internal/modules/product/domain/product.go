// @kthulu:domain:product
package domain

import (
	"time"
	
)

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Product represents a product entity
type Product struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Name string `json:"name" gorm:"name"`
	Price float64 `json:"price" gorm:"price"`
}

// TableName overrides the table name used by Product to `Products`
func (Product) TableName() string {
	return "Products"
}

// ProductRepository defines the repository interface
type ProductRepository interface {
	Create(entity *Product) error
	GetByID(id uint) (*Product, error)
	Update(entity *Product) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Product, error)
}

// ProductService defines the service interface  
type ProductService interface {
	CreateProduct(entity *Product) error
	GetProductByID(id uint) (*Product, error)
	UpdateProduct(entity *Product) error
	DeleteProduct(id uint) error
	ListProducts(filter SearchFilter) ([]*Product, error)
}
