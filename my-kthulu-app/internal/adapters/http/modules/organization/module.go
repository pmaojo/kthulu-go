// @kthulu:module:organization
// @kthulu:category:Custom
package organization

import (
"go.uber.org/fx"

"my-kthulu-app/internal/adapters/http/modules/organization/repository"
"my-kthulu-app/internal/adapters/http/modules/organization/service"
"my-kthulu-app/internal/adapters/http/modules/organization/handlers"
)

// Providers returns the Fx providers for the organization module
func Providers() fx.Option {
return fx.Options(
fx.Provide(
repository.NewOrganizationRepository,
service.NewOrganizationService,
handlers.NewOrganizationHandler,
),
)
}
