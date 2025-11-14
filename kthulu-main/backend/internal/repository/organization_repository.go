// @kthulu:module:org
package repository

import (
	"context"
	"time"

	"backend/internal/domain"
)

// OrganizationRepository defines behavior for organization persistence.
type OrganizationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, org *domain.Organization) error
	FindByID(ctx context.Context, id uint) (*domain.Organization, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	Update(ctx context.Context, org *domain.Organization) error
	Delete(ctx context.Context, id uint) error

	// Query operations
	List(ctx context.Context, limit, offset int) ([]*domain.Organization, error)
	Count(ctx context.Context) (int64, error)
	FindByDomain(ctx context.Context, domain string) (*domain.Organization, error)
	FindByOwner(ctx context.Context, userID uint) ([]*domain.Organization, error)

	// Existence checks
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ExistsByDomain(ctx context.Context, domain string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
}

// OrganizationUserRepository defines behavior for organization user relationship persistence.
type OrganizationUserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, orgUser *domain.OrganizationUser) error
	FindByID(ctx context.Context, id uint) (*domain.OrganizationUser, error)
	FindByOrganizationAndUser(ctx context.Context, organizationID, userID uint) (*domain.OrganizationUser, error)
	Update(ctx context.Context, orgUser *domain.OrganizationUser) error
	Delete(ctx context.Context, id uint) error

	// Query operations
	FindByOrganization(ctx context.Context, organizationID uint) ([]*domain.OrganizationUser, error)
	FindByUser(ctx context.Context, userID uint) ([]*domain.OrganizationUser, error)
	FindByRole(ctx context.Context, organizationID uint, role domain.OrganizationRole) ([]*domain.OrganizationUser, error)
	CountByOrganization(ctx context.Context, organizationID uint) (int64, error)

	// Permission checks
	IsUserInOrganization(ctx context.Context, organizationID, userID uint) (bool, error)
	GetUserRole(ctx context.Context, organizationID, userID uint) (domain.OrganizationRole, error)
	HasRole(ctx context.Context, organizationID, userID uint, role domain.OrganizationRole) (bool, error)

	// Bulk operations
	RemoveUserFromOrganization(ctx context.Context, organizationID, userID uint) error
	UpdateUserRole(ctx context.Context, organizationID, userID uint, role domain.OrganizationRole) error
}

// InvitationRepository defines behavior for invitation persistence.
type InvitationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, invitation *domain.Invitation) error
	FindByID(ctx context.Context, id uint) (*domain.Invitation, error)
	FindByToken(ctx context.Context, token string) (*domain.Invitation, error)
	Update(ctx context.Context, invitation *domain.Invitation) error
	Delete(ctx context.Context, id uint) error

	// Query operations
	FindByOrganization(ctx context.Context, organizationID uint) ([]*domain.Invitation, error)
	FindByEmail(ctx context.Context, email string) ([]*domain.Invitation, error)
	FindByInviter(ctx context.Context, inviterID uint) ([]*domain.Invitation, error)
	FindByStatus(ctx context.Context, status domain.InvitationStatus) ([]*domain.Invitation, error)
	FindExpired(ctx context.Context) ([]*domain.Invitation, error)

	// Existence checks
	ExistsByToken(ctx context.Context, token string) (bool, error)
	ExistsPendingByEmail(ctx context.Context, organizationID uint, email string) (bool, error)

	// Cleanup operations
	DeleteExpired(ctx context.Context, before time.Time) error
	MarkExpired(ctx context.Context, before time.Time) error
}
