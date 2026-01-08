// @kthulu:module:users
package modules

import (
	"go.uber.org/fx"

	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
	users "github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/users"
)

// UsersModule wires user auth components and HTTP handlers.
var UsersModule = fx.Options(
	fx.Provide(
		users.NewInMemoryUserRepository,
		users.NewAuthService,
		adapterhttp.NewUsersHandler,
	),
	fx.Invoke(func(h *adapterhttp.UsersHandler, rr *RouteRegistry) {
		rr.Register(h)
	}),
)
