// @kthulu:module:services
// @kthulu:generated:true
package services

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"github.com/example/workshop-api/internal/adapters/http/modules/services/handlers"
	"github.com/example/workshop-api/internal/adapters/http/modules/services/repository"
	"github.com/example/workshop-api/internal/adapters/http/modules/services/service"
)

// Providers returns the Fx providers for the services module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewServicesRepository,
                        service.NewServicesService,
                        handlers.NewServicesHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.ServicesHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
