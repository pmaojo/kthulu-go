// @kthulu:module:inventory
package inventory

import (
	"go.uber.org/fx"

	"backend/internal/infrastructure/db"
	"backend/internal/repository"
)

// Module provides fx.Options for inventory (stock) module.
// Includes inventory and warehouse management.
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			db.NewInventoryRepository,
			fx.As(new(repository.InventoryRepository)),
		),
	),
)
