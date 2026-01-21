// @kthulu:module:cache
// @kthulu:generated:true
package cache

import (
	"go.uber.org/fx"

	"feature-tour/internal/modules/cache/api"
	"feature-tour/internal/modules/cache/store"
	"feature-tour/internal/modules/cache/core"
)

// Providers returns the Fx providers for the cache module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewCacheRepository,
                        core.NewCacheService,
                        api.NewCacheHandler,
                ),
        )
}
