// @kthulu:module:projects
package repository

import (
	"context"

	"backend/internal/domain"
)

// ProjectRepository defines behavior for project persistence.
type ProjectRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, project *domain.Project) error
	FindByID(ctx context.Context, id uint) (*domain.Project, error)
	FindByName(ctx context.Context, name string) (*domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id uint) error

	// Query operations
	List(ctx context.Context, limit, offset int) ([]*domain.Project, error)
	Count(ctx context.Context) (int64, error)

	// Paginated operations
	FindPaginated(ctx context.Context, params PaginationParams) (PaginationResult[*domain.Project], error)
	SearchPaginated(ctx context.Context, query string, params PaginationParams) (PaginationResult[*domain.Project], error)

	// Existence checks
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
}
