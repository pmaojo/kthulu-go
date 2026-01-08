// @kthulu:domain:property
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

// Property represents a property entity
type Property struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Location    string  `json:"location"`
	HostID      uint    `json:"host_id"`
	ImageURL    string  `json:"image_url"`
}

// TableName overrides the table name used by Property to `Properties`
func (Property) TableName() string {
	return "Properties"
}

// PropertyRepository defines the repository interface
type PropertyRepository interface {
	Create(entity *Property) error
	GetByID(id uint) (*Property, error)
	Update(entity *Property) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Property, error)
}

// PropertyService defines the service interface  
type PropertyService interface {
	CreateProperty(entity *Property) error
	GetPropertyByID(id uint) (*Property, error)
	UpdateProperty(entity *Property) error
	DeleteProperty(id uint) error
	ListProperties(filter SearchFilter) ([]*Property, error)
}
