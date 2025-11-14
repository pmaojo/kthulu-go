// @kthulu:module:modules
package repository

import (
	"context"

	"backend/internal/domain"
)

// ModuleRepository defines behavior for module persistence.
type ModuleRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, module *domain.ModuleInfo) error
	FindByID(ctx context.Context, id uint) (*domain.ModuleInfo, error)
	FindByName(ctx context.Context, name string) (*domain.ModuleInfo, error)
	Update(ctx context.Context, module *domain.ModuleInfo) error
	Delete(ctx context.Context, id uint) error

	// Query operations
	List(ctx context.Context, limit, offset int) ([]*domain.ModuleInfo, error)
	ListByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.ModuleInfo, error)
	Count(ctx context.Context) (int64, error)

	// Paginated operations
	FindPaginated(ctx context.Context, params PaginationParams) (PaginationResult[*domain.ModuleInfo], error)
	SearchPaginated(ctx context.Context, query string, params PaginationParams) (PaginationResult[*domain.ModuleInfo], error)

	// Existence checks
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
}
