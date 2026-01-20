// @kthulu:module:product
// @kthulu:generated:true
package product

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"web-test/internal/modules/product/api"
	"web-test/internal/modules/product/store"
	"web-test/internal/modules/product/core"
)

// Providers returns the Fx providers for the product module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewProductRepository,
                        core.NewProductService,
                        api.NewProductHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *api.ProductHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
