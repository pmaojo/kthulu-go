// @kthulu:module:bookings
// @kthulu:generated:true
package bookings

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"github.com/example/workshop-api/internal/adapters/http/modules/bookings/handlers"
	"github.com/example/workshop-api/internal/adapters/http/modules/bookings/repository"
	"github.com/example/workshop-api/internal/adapters/http/modules/bookings/service"
)

// Providers returns the Fx providers for the bookings module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewBookingsRepository,
                        service.NewBookingsService,
                        handlers.NewBookingsHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.BookingsHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
