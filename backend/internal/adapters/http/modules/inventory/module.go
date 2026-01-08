// @kthulu:module:inventory
package inventory

import (
	"go.uber.org/fx"

	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"
	"github.com/pmaojo/kthulu-go/backend/internal/domain/repository"
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
