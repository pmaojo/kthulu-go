// @kthulu:core:organization
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

// Organization represents a organization entity
type Organization struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
}

// TableName overrides the table name used by Organization to `Organizations`
func (Organization) TableName() string {
	return "Organizations"
}

// OrganizationRepository defines the repository interface
type OrganizationRepository interface {
	Create(entity *Organization) error
	GetByID(id uint) (*Organization, error)
	Update(entity *Organization) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Organization, error)
}

// OrganizationService defines the service interface  
type OrganizationService interface {
	CreateOrganization(entity *Organization) error
	GetOrganizationByID(id uint) (*Organization, error)
	UpdateOrganization(entity *Organization) error
	DeleteOrganization(id uint) error
	ListOrganizations(filter SearchFilter) ([]*Organization, error)
}
