// @kthulu:module:orders
// @kthulu:generated:true
package orders

import (
	"go.uber.org/fx"
	"shop/internal/adapters/http/modules/orders/handlers"
	"shop/internal/adapters/http/modules/orders/repository"
	"shop/internal/adapters/http/modules/orders/service"
)

// Providers returns the Fx providers for the orders module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewOrdersRepository,
                        service.NewOrdersService,
                        handlers.NewOrdersHandler,
                ),
        )
}
