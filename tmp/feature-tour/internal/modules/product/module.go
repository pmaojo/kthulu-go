// @kthulu:module:product
// @kthulu:generated:true
package product

import (
	"go.uber.org/fx"

	"feature-tour/internal/modules/product/api"
	"feature-tour/internal/modules/product/store"
	"feature-tour/internal/modules/product/core"
)

// Providers returns the Fx providers for the product module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewProductRepository,
                        core.NewProductService,
                        api.NewProductHandler,
                ),
        )
}
