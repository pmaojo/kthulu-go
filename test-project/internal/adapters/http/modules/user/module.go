// @kthulu:module:user
// @kthulu:category:Custom
package user

import (
"go.uber.org/fx"

"test-project/internal/adapters/http/modules/user/repository"
"test-project/internal/adapters/http/modules/user/service"
"test-project/internal/adapters/http/modules/user/handlers"
)

// Providers returns the Fx providers for the user module
func Providers() fx.Option {
return fx.Options(
fx.Provide(
repository.NewUserRepository,
service.NewUserService,
handlers.NewUserHandler,
),
)
}
