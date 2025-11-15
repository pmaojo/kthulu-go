// @kthulu:module:product
package product

import (
	"go.uber.org/fx"

	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// Module provides fx.Options for product (catalog) module.
// Includes product catalog management.
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			db.NewProductRepository,
			fx.As(new(repository.ProductRepository)),
		),
	),
)
