// @kthulu:module:booking
// @kthulu:generated:true
package booking

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"airbnb-v2/internal/modules/booking/handlers"
	"airbnb-v2/internal/modules/booking/repository"
	"airbnb-v2/internal/modules/booking/service"
)

// Providers returns the Fx providers for the booking module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewBookingRepository,
                        service.NewBookingService,
                        handlers.NewBookingHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.BookingHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
