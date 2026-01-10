// @kthulu:domain:auth
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

// Auth represents a auth entity
type Auth struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

}

// TableName overrides the table name used by Auth to `Auths`
func (Auth) TableName() string {
	return "Auths"
}

// AuthRepository defines the repository interface
type AuthRepository interface {
	Create(entity *Auth) error
	GetByID(id uint) (*Auth, error)
	Update(entity *Auth) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Auth, error)
}

// AuthService defines the service interface
type AuthService interface {
	CreateAuth(entity *Auth) error
	GetAuthByID(id uint) (*Auth, error)
	UpdateAuth(entity *Auth) error
	DeleteAuth(id uint) error
	ListAuths(filter SearchFilter) ([]*Auth, error)
}
