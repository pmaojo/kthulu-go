// @kthulu:module:templates
package repository

import (
	"context"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
)

// TemplateRepository defines behavior for template persistence.
type TemplateRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, template *domain.Template) error
	FindByID(ctx context.Context, id uint) (*domain.Template, error)
	FindByName(ctx context.Context, name string) (*domain.Template, error)
	Update(ctx context.Context, template *domain.Template) error
	Delete(ctx context.Context, id uint) error

	// Query operations
	List(ctx context.Context, limit, offset int) ([]*domain.Template, error)
	ListByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.Template, error)
	Count(ctx context.Context) (int64, error)

	// Paginated operations
	FindPaginated(ctx context.Context, params PaginationParams) (PaginationResult[*domain.Template], error)
	SearchPaginated(ctx context.Context, query string, params PaginationParams) (PaginationResult[*domain.Template], error)

	// Existence checks
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
}

// TemplateRegistryRepository defines behavior for template registry persistence.
type TemplateRegistryRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, registry *domain.TemplateRegistry) error
	FindByID(ctx context.Context, id uint) (*domain.TemplateRegistry, error)
	FindByName(ctx context.Context, name string) (*domain.TemplateRegistry, error)
	Update(ctx context.Context, registry *domain.TemplateRegistry) error
	Delete(ctx context.Context, id uint) error

	// Query operations
	List(ctx context.Context, limit, offset int) ([]*domain.TemplateRegistry, error)
	Count(ctx context.Context) (int64, error)

	// Existence checks
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
}
