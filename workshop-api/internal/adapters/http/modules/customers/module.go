// @kthulu:module:customers
// @kthulu:generated:true
package customers

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"github.com/example/workshop-api/internal/adapters/http/modules/customers/handlers"
	"github.com/example/workshop-api/internal/adapters/http/modules/customers/repository"
	"github.com/example/workshop-api/internal/adapters/http/modules/customers/service"
)

// Providers returns the Fx providers for the customers module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewCustomersRepository,
                        service.NewCustomersService,
                        handlers.NewCustomersHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.CustomersHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
