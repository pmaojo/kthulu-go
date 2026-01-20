// @kthulu:module:organization
// @kthulu:generated:true
package organization

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"web-test/internal/modules/organization/api"
	"web-test/internal/modules/organization/store"
	"web-test/internal/modules/organization/core"
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
