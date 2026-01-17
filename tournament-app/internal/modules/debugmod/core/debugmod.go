// @kthulu:core:debugmod
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

// Debugmod represents a debugmod entity
type Debugmod struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
}

// TableName overrides the table name used by Debugmod to `Debugmods`
func (Debugmod) TableName() string {
	return "Debugmods"
}

// DebugmodRepository defines the repository interface
type DebugmodRepository interface {
	Create(entity *Debugmod) error
	GetByID(id uint) (*Debugmod, error)
	Update(entity *Debugmod) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Debugmod, error)
}

// DebugmodService defines the service interface  
type DebugmodService interface {
	CreateDebugmod(entity *Debugmod) error
	GetDebugmodByID(id uint) (*Debugmod, error)
	UpdateDebugmod(entity *Debugmod) error
	DeleteDebugmod(id uint) error
	ListDebugmods(filter SearchFilter) ([]*Debugmod, error)
}
