// @kthulu:module:contacts
package repository

import (
	"context"

	"backend/internal/domain"
)

// ContactRepository defines the interface for contact data operations
type ContactRepository interface {
	// Contact operations
	Create(ctx context.Context, contact *domain.Contact) error
	GetByID(ctx context.Context, organizationID, contactID uint) (*domain.Contact, error)
	GetByEmail(ctx context.Context, organizationID uint, email string) (*domain.Contact, error)
	Update(ctx context.Context, contact *domain.Contact) error
	Delete(ctx context.Context, organizationID, contactID uint) error
	List(ctx context.Context, organizationID uint, filters ContactFilters) ([]*domain.Contact, int64, error)

	// Address operations
	CreateAddress(ctx context.Context, address *domain.ContactAddress) error
	GetAddressesByContactID(ctx context.Context, contactID uint) ([]*domain.ContactAddress, error)
	GetAddressByID(ctx context.Context, contactID, addressID uint) (*domain.ContactAddress, error)
	UpdateAddress(ctx context.Context, address *domain.ContactAddress) error
	DeleteAddress(ctx context.Context, contactID, addressID uint) error
	SetPrimaryAddress(ctx context.Context, contactID, addressID uint) error

	// Phone operations
	CreatePhone(ctx context.Context, phone *domain.ContactPhone) error
	GetPhonesByContactID(ctx context.Context, contactID uint) ([]*domain.ContactPhone, error)
	GetPhoneByID(ctx context.Context, contactID, phoneID uint) (*domain.ContactPhone, error)
	UpdatePhone(ctx context.Context, phone *domain.ContactPhone) error
	DeletePhone(ctx context.Context, contactID, phoneID uint) error
	SetPrimaryPhone(ctx context.Context, contactID, phoneID uint) error

	// Bulk operations
	BulkCreate(ctx context.Context, contacts []*domain.Contact) error
	BulkUpdate(ctx context.Context, contacts []*domain.Contact) error
	BulkDelete(ctx context.Context, organizationID uint, contactIDs []uint) error

	// Statistics
	GetContactStats(ctx context.Context, organizationID uint) (*ContactStats, error)
}

// ContactFilters represents filters for contact listing
type ContactFilters struct {
	Type        domain.ContactType `json:"type,omitempty"`
	IsActive    *bool              `json:"isActive,omitempty"`
	Search      string             `json:"search,omitempty"`      // Search in name, email, company
	CreatedFrom *string            `json:"createdFrom,omitempty"` // ISO date string
	CreatedTo   *string            `json:"createdTo,omitempty"`   // ISO date string

	// Pagination
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"pageSize" validate:"min=1,max=100"`

	// Sorting
	SortBy    string `json:"sortBy,omitempty"`    // name, email, company_name, created_at, updated_at
	SortOrder string `json:"sortOrder,omitempty"` // asc, desc
}

// ContactStats represents contact statistics for an organization
type ContactStats struct {
	TotalContacts    int64 `json:"totalContacts"`
	ActiveContacts   int64 `json:"activeContacts"`
	InactiveContacts int64 `json:"inactiveContacts"`
	CustomerCount    int64 `json:"customerCount"`
	SupplierCount    int64 `json:"supplierCount"`
	LeadCount        int64 `json:"leadCount"`
	PartnerCount     int64 `json:"partnerCount"`
	RecentContacts   int64 `json:"recentContacts"` // Contacts created in last 30 days
}

// DefaultContactFilters returns default filters for contact listing
func DefaultContactFilters() ContactFilters {
	active := true
	return ContactFilters{
		IsActive:  &active,
		Page:      1,
		PageSize:  20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// Validate validates the contact filters
func (f *ContactFilters) Validate() error {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		f.SortOrder = "desc"
	}
	return nil
}

// GetOffset returns the offset for pagination
func (f *ContactFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}
