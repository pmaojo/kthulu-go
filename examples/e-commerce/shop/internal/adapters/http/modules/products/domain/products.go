// @kthulu:domain:products
package domain

import "time"

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Products represents a products entity
type Products struct {
        ID        uint      `json:"id" gorm:"primaryKey"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`

        Name string `json:"name" gorm:"name"`
        Description string `json:"description" gorm:"description"`
        Price float64 `json:"price" gorm:"price"`
        Stock int `json:"stock" gorm:"stock"`

}

// TableName overrides the table name used by User to `productss`
func (Products) TableName() string {
	return "productss"
}

// ProductsRepository defines the repository interface
type ProductsRepository interface {
        Create(entity *Products) error
        GetByID(id uint) (*Products, error)
        Update(entity *Products) error
        Delete(id uint) error
        List(filter SearchFilter) ([]*Products, error)
}

// ProductsService defines the service interface
type ProductsService interface {
        CreateProducts(entity *Products) error
        GetProductsByID(id uint) (*Products, error)
        UpdateProducts(entity *Products) error
        DeleteProducts(id uint) error
        ListProductss(filter SearchFilter) ([]*Products, error)
}
