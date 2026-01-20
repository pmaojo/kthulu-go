// @kthulu:module:organization
// @kthulu:generated:true
package organization

import (
	"go.uber.org/fx"

	"kitchen-sink-test/internal/modules/organization/api"
	"kitchen-sink-test/internal/modules/organization/store"
	"kitchen-sink-test/internal/modules/organization/core"
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
