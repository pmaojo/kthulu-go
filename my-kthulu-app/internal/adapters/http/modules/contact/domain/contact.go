// @kthulu:domain:contact
package domain

import "time"

// Contact represents a contact entity
type Contact struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Add your fields here
}

// ContactRepository defines the repository interface
type ContactRepository interface {
	Create(entity *Contact) error
	GetByID(id uint) (*Contact, error)
	Update(entity *Contact) error
	Delete(id uint) error
	List() ([]*Contact, error)
}

// ContactService defines the service interface
type ContactService interface {
	CreateContact(entity *Contact) error
	GetContactByID(id uint) (*Contact, error)
	UpdateContact(entity *Contact) error
	DeleteContact(id uint) error
	ListContacts() ([]*Contact, error)
}
