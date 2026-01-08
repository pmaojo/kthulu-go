// @kthulu:module:product
// @kthulu:generated:true
package product

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"demo-app/internal/modules/product/handlers"
	"demo-app/internal/modules/product/repository"
	"demo-app/internal/modules/product/service"
)

// Providers returns the Fx providers for the product module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewProductRepository,
                        service.NewProductService,
                        handlers.NewProductHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.ProductHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
