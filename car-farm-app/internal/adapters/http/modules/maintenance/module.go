// @kthulu:module:maintenance
// @kthulu:generated:true
package maintenance

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"car-farm-app/internal/adapters/http/modules/maintenance/handlers"
	"car-farm-app/internal/adapters/http/modules/maintenance/repository"
	"car-farm-app/internal/adapters/http/modules/maintenance/service"
)

// Providers returns the Fx providers for the maintenance module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewMaintenanceRepository,
                        service.NewMaintenanceService,
                        handlers.NewMaintenanceHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.MaintenanceHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
