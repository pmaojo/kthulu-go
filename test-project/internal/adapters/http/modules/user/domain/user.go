// @kthulu:domain:user
package domain

import "time"

// User represents a user entity
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Add your fields here
}

// UserRepository defines the repository interface
type UserRepository interface {
	Create(entity *User) error
	GetByID(id uint) (*User, error)
	Update(entity *User) error
	Delete(id uint) error
	List() ([]*User, error)
}

// UserService defines the service interface
type UserService interface {
	CreateUser(entity *User) error
	GetUserByID(id uint) (*User, error)
	UpdateUser(entity *User) error
	DeleteUser(id uint) error
	ListUsers() ([]*User, error)
}
