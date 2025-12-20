// @kthulu:domain:payments
package domain

import "time"

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Payments represents a payments entity
type Payments struct {
        ID        uint      `json:"id" gorm:"primaryKey"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`

        Amount float64 `json:"amount" gorm:"amount"`
        Provider string `json:"provider" gorm:"provider"`
        Transaction_id string `json:"transaction_id" gorm:"transaction_id"`

}

// TableName overrides the table name used by User to `paymentss`
func (Payments) TableName() string {
	return "paymentss"
}

// PaymentsRepository defines the repository interface
type PaymentsRepository interface {
        Create(entity *Payments) error
        GetByID(id uint) (*Payments, error)
        Update(entity *Payments) error
        Delete(id uint) error
        List(filter SearchFilter) ([]*Payments, error)
}

// PaymentsService defines the service interface
type PaymentsService interface {
        CreatePayments(entity *Payments) error
        GetPaymentsByID(id uint) (*Payments, error)
        UpdatePayments(entity *Payments) error
        DeletePayments(id uint) error
        ListPaymentss(filter SearchFilter) ([]*Payments, error)
}
