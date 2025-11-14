// @kthulu:module:product
package product

import (
	"go.uber.org/fx"

	"backend/internal/infrastructure/db"
	"backend/internal/repository"
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
