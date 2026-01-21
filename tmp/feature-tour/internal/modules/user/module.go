// @kthulu:module:user
// @kthulu:generated:true
package user

import (
	"go.uber.org/fx"

	"feature-tour/internal/modules/user/api"
	"feature-tour/internal/modules/user/store"
	"feature-tour/internal/modules/user/core"
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
