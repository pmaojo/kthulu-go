// @kthulu:domain:auth
package domain

import "time"

// Auth represents a auth entity
type Auth struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Add your fields here
}

// AuthRepository defines the repository interface
type AuthRepository interface {
	Create(entity *Auth) error
	GetByID(id uint) (*Auth, error)
	Update(entity *Auth) error
	Delete(id uint) error
	List() ([]*Auth, error)
}

// AuthService defines the service interface
type AuthService interface {
	CreateAuth(entity *Auth) error
	GetAuthByID(id uint) (*Auth, error)
	UpdateAuth(entity *Auth) error
	DeleteAuth(id uint) error
	ListAuths() ([]*Auth, error)
}
