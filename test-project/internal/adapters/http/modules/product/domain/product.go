// @kthulu:domain:product
package domain

import "time"

// Product represents a product entity
type Product struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Add your fields here
}

// ProductRepository defines the repository interface
type ProductRepository interface {
	Create(entity *Product) error
	GetByID(id uint) (*Product, error)
	Update(entity *Product) error
	Delete(id uint) error
	List() ([]*Product, error)
}

// ProductService defines the service interface
type ProductService interface {
	CreateProduct(entity *Product) error
	GetProductByID(id uint) (*Product, error)
	UpdateProduct(entity *Product) error
	DeleteProduct(id uint) error
	ListProducts() ([]*Product, error)
}
