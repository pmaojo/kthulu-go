// @kthulu:module:user
// @kthulu:category:Custom
package user

import (
"go.uber.org/fx"

"shop/internal/adapters/http/modules/user/repository"
"shop/internal/adapters/http/modules/user/service"
"shop/internal/adapters/http/modules/user/handlers"
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
