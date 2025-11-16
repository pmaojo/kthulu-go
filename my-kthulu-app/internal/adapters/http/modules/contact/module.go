// @kthulu:module:contact
// @kthulu:category:Custom
package contact

import (
"go.uber.org/fx"

"my-kthulu-app/internal/adapters/http/modules/contact/repository"
"my-kthulu-app/internal/adapters/http/modules/contact/service"
"my-kthulu-app/internal/adapters/http/modules/contact/handlers"
)

// Providers returns the Fx providers for the contact module
func Providers() fx.Option {
return fx.Options(
fx.Provide(
repository.NewContactRepository,
service.NewContactService,
handlers.NewContactHandler,
),
)
}
