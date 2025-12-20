// @kthulu:module:products
// @kthulu:generated:true
package products

import (
	"go.uber.org/fx"
	"shop/internal/adapters/http/modules/products/handlers"
	"shop/internal/adapters/http/modules/products/repository"
	"shop/internal/adapters/http/modules/products/service"
)

// Providers returns the Fx providers for the products module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewProductsRepository,
                        service.NewProductsService,
                        handlers.NewProductsHandler,
                ),
        )
}
