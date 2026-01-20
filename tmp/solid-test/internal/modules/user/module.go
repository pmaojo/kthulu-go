// @kthulu:module:user
// @kthulu:generated:true
package user

import (
	"go.uber.org/fx"

	"solid-test/internal/modules/user/api"
	"solid-test/internal/modules/user/store"
	"solid-test/internal/modules/user/core"
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
