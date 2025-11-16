// @kthulu:module:auth
// @kthulu:category:Custom
package auth

import (
"go.uber.org/fx"

"test-project/internal/adapters/http/modules/auth/repository"
"test-project/internal/adapters/http/modules/auth/service"
"test-project/internal/adapters/http/modules/auth/handlers"
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
