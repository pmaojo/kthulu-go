// @kthulu:domain:user
package domain

import (
	"time"
	
)

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// User represents a user entity
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
}

// TableName overrides the table name used by User to `Users`
func (User) TableName() string {
	return "Users"
}

// UserRepository defines the repository interface
type UserRepository interface {
	Create(entity *User) error
	GetByID(id uint) (*User, error)
	Update(entity *User) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*User, error)
}

// UserService defines the service interface  
type UserService interface {
	CreateUser(entity *User) error
	GetUserByID(id uint) (*User, error)
	UpdateUser(entity *User) error
	DeleteUser(id uint) error
	ListUsers(filter SearchFilter) ([]*User, error)
}
