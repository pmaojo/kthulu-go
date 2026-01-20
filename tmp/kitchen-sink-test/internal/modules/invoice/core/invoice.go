// @kthulu:core:invoice
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

// Invoice represents a invoice entity
type Invoice struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Customer_id int `json:"customer_id,string" gorm:"customer_id"`
	Amount int `json:"amount,string" gorm:"amount"`
	Status string `json:"status" gorm:"status"`
	Due_date string `json:"due_date" gorm:"due_date"`
}

// TableName overrides the table name used by Invoice to `Invoices`
func (Invoice) TableName() string {
	return "Invoices"
}

// InvoiceRepository defines the repository interface
type InvoiceRepository interface {
	Create(entity *Invoice) error
	GetByID(id uint) (*Invoice, error)
	Update(entity *Invoice) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Invoice, error)
}

// InvoiceService defines the service interface  
type InvoiceService interface {
	CreateInvoice(entity *Invoice) error
	GetInvoiceByID(id uint) (*Invoice, error)
	UpdateInvoice(entity *Invoice) error
	DeleteInvoice(id uint) error
	ListInvoices(filter SearchFilter) ([]*Invoice, error)
}
