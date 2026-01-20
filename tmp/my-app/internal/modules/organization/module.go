// @kthulu:module:organization
// @kthulu:generated:true
package organization

import (
	"go.uber.org/fx"

	"my-app/internal/modules/organization/api"
	"my-app/internal/modules/organization/store"
	"my-app/internal/modules/organization/core"
)

// Providers returns the Fx providers for the organization module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewOrganizationRepository,
                        core.NewOrganizationService,
                        api.NewOrganizationHandler,
                ),
        )
}
