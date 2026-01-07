// @kthulu:domain:users
package domain

import "time"

// Users represents a users entity
type Users struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Add your fields here
}

// UsersRepository defines the repository interface
type UsersRepository interface {
	Create(entity *Users) error
	GetByID(id uint) (*Users, error)
	Update(entity *Users) error
	Delete(id uint) error
	List() ([]*Users, error)
}

// UsersService defines the service interface
type UsersService interface {
	CreateUsers(entity *Users) error
	GetUsersByID(id uint) (*Users, error)
	UpdateUsers(entity *Users) error
	DeleteUsers(id uint) error
	ListUserses() ([]*Users, error)
}
