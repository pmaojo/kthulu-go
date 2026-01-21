// @kthulu:module:storage
// @kthulu:generated:true
package storage

import (
	"go.uber.org/fx"

	"feature-tour/internal/modules/storage/api"
	"feature-tour/internal/modules/storage/store"
	"feature-tour/internal/modules/storage/core"
)

// Providers returns the Fx providers for the storage module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewStorageRepository,
                        core.NewStorageService,
                        api.NewStorageHandler,
                ),
        )
}
