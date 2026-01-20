// @kthulu:module:organization
// @kthulu:generated:true
package organization

import (
	"go.uber.org/fx"

	"solid-test/internal/modules/organization/api"
	"solid-test/internal/modules/organization/store"
	"solid-test/internal/modules/organization/core"
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
