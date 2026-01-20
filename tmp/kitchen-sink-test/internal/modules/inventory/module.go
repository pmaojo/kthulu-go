// @kthulu:module:inventory
// @kthulu:generated:true
package inventory

import (
	"go.uber.org/fx"

	"kitchen-sink-test/internal/modules/inventory/api"
	"kitchen-sink-test/internal/modules/inventory/store"
	"kitchen-sink-test/internal/modules/inventory/core"
)

// Providers returns the Fx providers for the inventory module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewInventoryRepository,
                        core.NewInventoryService,
                        api.NewInventoryHandler,
                ),
        )
}
