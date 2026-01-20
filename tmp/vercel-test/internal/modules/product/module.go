// @kthulu:module:product
// @kthulu:generated:true
package product

import (
	"go.uber.org/fx"

	"vercel-test/internal/modules/product/api"
	"vercel-test/internal/modules/product/store"
	"vercel-test/internal/modules/product/core"
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
