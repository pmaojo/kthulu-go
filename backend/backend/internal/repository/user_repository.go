// @kthulu:module:auth
package repository

import (
	"context"
	"time"

	"github.com/kthulu/kthulu-go/backend/internal/domain"
)

// UserRepository defines behavior for user persistence.
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uint) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uint) error

	// Query operations
	List(ctx context.Context, limit, offset int) ([]*domain.User, error)
	Count(ctx context.Context) (int64, error)
	FindByRole(ctx context.Context, roleID uint) ([]*domain.User, error)
	FindUnconfirmed(ctx context.Context, olderThan *time.Time) ([]*domain.User, error)

	// Paginated operations
	FindPaginated(ctx context.Context, params PaginationParams) (PaginationResult[*domain.User], error)
	SearchPaginated(ctx context.Context, query string, params PaginationParams) (PaginationResult[*domain.User], error)

	// Existence checks
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
}
