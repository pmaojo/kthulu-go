// @kthulu:module:users
// @kthulu:category:Custom
package users

import (
"go.uber.org/fx"

"github.com/example/workshop-api/internal/adapters/http/modules/users/repository"
"github.com/example/workshop-api/internal/adapters/http/modules/users/service"
"github.com/example/workshop-api/internal/adapters/http/modules/users/handlers"
)

// Providers returns the Fx providers for the users module
func Providers() fx.Option {
return fx.Options(
fx.Provide(
repository.NewUsersRepository,
service.NewUsersService,
handlers.NewUsersHandler,
),
)
}
