// @kthulu:module:organization
// @kthulu:generated:true
package organization

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"gth-test-app/internal/modules/organization/api"
	"gth-test-app/internal/modules/organization/store"
	"gth-test-app/internal/modules/organization/core"
)

// Providers returns the Fx providers for the organization module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewOrganizationRepository,
                        core.NewOrganizationService,
                        api.NewOrganizationHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *api.OrganizationHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
