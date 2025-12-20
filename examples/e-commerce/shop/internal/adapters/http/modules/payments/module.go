// @kthulu:module:payments
// @kthulu:generated:true
package payments

import (
	"go.uber.org/fx"
	"shop/internal/adapters/http/modules/payments/handlers"
	"shop/internal/adapters/http/modules/payments/repository"
	"shop/internal/adapters/http/modules/payments/service"
)

// Providers returns the Fx providers for the payments module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewPaymentsRepository,
                        service.NewPaymentsService,
                        handlers.NewPaymentsHandler,
                ),
        )
}
