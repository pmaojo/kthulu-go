// @kthulu:module:cars
// @kthulu:generated:true
package cars

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"github.com/example/workshop-api/internal/adapters/http/modules/cars/handlers"
	"github.com/example/workshop-api/internal/adapters/http/modules/cars/repository"
	"github.com/example/workshop-api/internal/adapters/http/modules/cars/service"
)

// Providers returns the Fx providers for the cars module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewCarsRepository,
                        service.NewCarsService,
                        handlers.NewCarsHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.CarsHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
