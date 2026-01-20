// @kthulu:core:contact
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

// Contact represents a contact entity
type Contact struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Name string `json:"name" gorm:"name"`
	Email string `json:"email" gorm:"email"`
	Phone string `json:"phone" gorm:"phone"`
	Company string `json:"company" gorm:"company"`
}

// TableName overrides the table name used by Contact to `Contacts`
func (Contact) TableName() string {
	return "Contacts"
}

// ContactRepository defines the repository interface
type ContactRepository interface {
	Create(entity *Contact) error
	GetByID(id uint) (*Contact, error)
	Update(entity *Contact) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Contact, error)
}

// ContactService defines the service interface  
type ContactService interface {
	CreateContact(entity *Contact) error
	GetContactByID(id uint) (*Contact, error)
	UpdateContact(entity *Contact) error
	DeleteContact(id uint) error
	ListContacts(filter SearchFilter) ([]*Contact, error)
}
