// @kthulu:module:access
package repository

import (
	"context"

	"github.com/kthulu/kthulu-go/backend/internal/domain"
)

// RoleRepository defines behavior for role persistence.
type RoleRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, role *domain.Role) error
	FindByID(ctx context.Context, id uint) (*domain.Role, error)
	FindByUserID(ctx context.Context, userID uint) (*domain.Role, error) // Load a role associated with a user
	FindByName(ctx context.Context, name string) (*domain.Role, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id uint) error

	// Query operations
	List(ctx context.Context) ([]*domain.Role, error)
	Count(ctx context.Context) (int64, error)

	// Existence checks
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)

	// Permission operations
	AddPermission(ctx context.Context, roleID, permissionID uint) error
	RemovePermission(ctx context.Context, roleID, permissionID uint) error
	GetRolePermissions(ctx context.Context, roleID uint) ([]*domain.Permission, error)
}

// PermissionRepository defines behavior for permission persistence.
type PermissionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, permission *domain.Permission) error
	FindByID(ctx context.Context, id uint) (*domain.Permission, error)
	FindByResourceAndAction(ctx context.Context, resource, action string) (*domain.Permission, error)
	Update(ctx context.Context, permission *domain.Permission) error
	Delete(ctx context.Context, id uint) error

	// Query operations
	List(ctx context.Context) ([]*domain.Permission, error)
	FindByResource(ctx context.Context, resource string) ([]*domain.Permission, error)
	Count(ctx context.Context) (int64, error)

	// Existence checks
	ExistsByResourceAndAction(ctx context.Context, resource, action string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
}
