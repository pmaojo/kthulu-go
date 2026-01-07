// @kthulu:domain:customers
package domain

import "time"

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Customers represents a customers entity
type Customers struct {
        ID        uint      `json:"id" gorm:"primaryKey"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`

        Name string `json:"name" gorm:"name"`
        Email string `json:"email" gorm:"email"`
        Phone string `json:"phone" gorm:"phone"`
        Address string `json:"address" gorm:"address"`

}

// TableName overrides the table name used by User to `customerss`
func (Customers) TableName() string {
	return "customerss"
}

// CustomersRepository defines the repository interface
type CustomersRepository interface {
        Create(entity *Customers) error
        GetByID(id uint) (*Customers, error)
        Update(entity *Customers) error
        Delete(id uint) error
        List(filter SearchFilter) ([]*Customers, error)
}

// CustomersService defines the service interface
type CustomersService interface {
        CreateCustomers(entity *Customers) error
        GetCustomersByID(id uint) (*Customers, error)
        UpdateCustomers(entity *Customers) error
        DeleteCustomers(id uint) error
        ListCustomerss(filter SearchFilter) ([]*Customers, error)
}
