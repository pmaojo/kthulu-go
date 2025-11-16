// @kthulu:domain:organization
package domain

import "time"

// Organization represents a organization entity
type Organization struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Add your fields here
}

// OrganizationRepository defines the repository interface
type OrganizationRepository interface {
	Create(entity *Organization) error
	GetByID(id uint) (*Organization, error)
	Update(entity *Organization) error
	Delete(id uint) error
	List() ([]*Organization, error)
}

// OrganizationService defines the service interface
type OrganizationService interface {
	CreateOrganization(entity *Organization) error
	GetOrganizationByID(id uint) (*Organization, error)
	UpdateOrganization(entity *Organization) error
	DeleteOrganization(id uint) error
	ListOrganizations() ([]*Organization, error)
}
