// @kthulu:module:product
// @kthulu:category:Custom
package product

import (
"go.uber.org/fx"

"my-kthulu-app/internal/adapters/http/modules/product/repository"
"my-kthulu-app/internal/adapters/http/modules/product/service"
"my-kthulu-app/internal/adapters/http/modules/product/handlers"
)

// Providers returns the Fx providers for the product module
func Providers() fx.Option {
return fx.Options(
fx.Provide(
repository.NewProductRepository,
service.NewProductService,
handlers.NewProductHandler,
),
)
}
