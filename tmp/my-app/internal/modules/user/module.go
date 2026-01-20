// @kthulu:module:user
// @kthulu:generated:true
package user

import (
	"go.uber.org/fx"

	"my-app/internal/modules/user/api"
	"my-app/internal/modules/user/store"
	"my-app/internal/modules/user/core"
)

// Providers returns the Fx providers for the user module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewUserRepository,
                        core.NewUserService,
                        api.NewUserHandler,
                ),
        )
}
