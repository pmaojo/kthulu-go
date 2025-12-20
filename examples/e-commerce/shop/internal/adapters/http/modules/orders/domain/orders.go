// @kthulu:domain:orders
package domain

import "time"

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Orders represents a orders entity
type Orders struct {
        ID        uint      `json:"id" gorm:"primaryKey"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`

        Total float64 `json:"total" gorm:"total"`
        Status string `json:"status" gorm:"status"`
        Customer_name string `json:"customer_name" gorm:"customer_name"`

}

// TableName overrides the table name used by User to `orderss`
func (Orders) TableName() string {
	return "orderss"
}

// OrdersRepository defines the repository interface
type OrdersRepository interface {
        Create(entity *Orders) error
        GetByID(id uint) (*Orders, error)
        Update(entity *Orders) error
        Delete(id uint) error
        List(filter SearchFilter) ([]*Orders, error)
}

// OrdersService defines the service interface
type OrdersService interface {
        CreateOrders(entity *Orders) error
        GetOrdersByID(id uint) (*Orders, error)
        UpdateOrders(entity *Orders) error
        DeleteOrders(id uint) error
        ListOrderss(filter SearchFilter) ([]*Orders, error)
}
