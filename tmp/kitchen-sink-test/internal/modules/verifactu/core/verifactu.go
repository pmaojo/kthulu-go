// @kthulu:core:verifactu
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

// Verifactu represents a verifactu entity
type Verifactu struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Invoice_id int `json:"invoice_id,string" gorm:"invoice_id"`
	Status string `json:"status" gorm:"status"`
	Fiscal_data string `json:"fiscal_data" gorm:"fiscal_data"`
}

// TableName overrides the table name used by Verifactu to `Verifactus`
func (Verifactu) TableName() string {
	return "Verifactus"
}

// VerifactuRepository defines the repository interface
type VerifactuRepository interface {
	Create(entity *Verifactu) error
	GetByID(id uint) (*Verifactu, error)
	Update(entity *Verifactu) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Verifactu, error)
}

// VerifactuService defines the service interface  
type VerifactuService interface {
	CreateVerifactu(entity *Verifactu) error
	GetVerifactuByID(id uint) (*Verifactu, error)
	UpdateVerifactu(entity *Verifactu) error
	DeleteVerifactu(id uint) error
	ListVerifactus(filter SearchFilter) ([]*Verifactu, error)
}
