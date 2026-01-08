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
	
	Title string `json:"title" gorm:"title"`
	Description string `json:"description" gorm:"description"`
	Price string `json:"price" gorm:"price"`
	Location string `json:"location" gorm:"location"`
	HostID string `json:"host_i_d" gorm:"host_i_d"`
	ImageURL string `json:"image_u_r_l" gorm:"image_u_r_l"`
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
