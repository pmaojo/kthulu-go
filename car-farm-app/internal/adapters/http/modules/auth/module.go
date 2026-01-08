// @kthulu:module:auth
// @kthulu:category:Custom
package auth

import (
"go.uber.org/fx"

"car-farm-app/internal/adapters/http/modules/auth/repository"
"car-farm-app/internal/adapters/http/modules/auth/service"
"car-farm-app/internal/adapters/http/modules/auth/handlers"
)

// Providers returns the Fx providers for the auth module
func Providers() fx.Option {
return fx.Options(
fx.Provide(
repository.NewAuthRepository,
service.NewAuthService,
handlers.NewAuthHandler,
),
)
}
