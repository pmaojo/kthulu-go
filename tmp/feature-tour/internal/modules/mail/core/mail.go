// @kthulu:core:mail
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

// Mail represents a mail entity
type Mail struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Name string `json:"name" gorm:"name"`
}

// TableName overrides the table name used by Mail to `Mails`
func (Mail) TableName() string {
	return "Mails"
}

// MailRepository defines the repository interface
type MailRepository interface {
	Create(entity *Mail) error
	GetByID(id uint) (*Mail, error)
	Update(entity *Mail) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Mail, error)
}

// MailService defines the service interface  
type MailService interface {
	CreateMail(entity *Mail) error
	GetMailByID(id uint) (*Mail, error)
	UpdateMail(entity *Mail) error
	DeleteMail(id uint) error
	ListMails(filter SearchFilter) ([]*Mail, error)
}
