// @kthulu:module:user
// @kthulu:generated:true
package user

import (
	"go.uber.org/fx"

	"kitchen-sink-test/internal/modules/user/api"
	"kitchen-sink-test/internal/modules/user/store"
	"kitchen-sink-test/internal/modules/user/core"
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
