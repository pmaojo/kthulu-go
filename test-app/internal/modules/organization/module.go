// @kthulu:module:organization
// @kthulu:generated:true
package organization

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"test-app/internal/modules/organization/handlers"
	"test-app/internal/modules/organization/repository"
	"test-app/internal/modules/organization/service"
)

// Providers returns the Fx providers for the organization module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewOrganizationRepository,
                        service.NewOrganizationService,
                        handlers.NewOrganizationHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.OrganizationHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
