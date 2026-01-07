// @kthulu:module:farms
// @kthulu:generated:true
package farms

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"car-farm-app/internal/adapters/http/modules/farms/handlers"
	"car-farm-app/internal/adapters/http/modules/farms/repository"
	"car-farm-app/internal/adapters/http/modules/farms/service"
)

// Providers returns the Fx providers for the farms module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewFarmsRepository,
                        service.NewFarmsService,
                        handlers.NewFarmsHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.FarmsHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
