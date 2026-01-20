// @kthulu:module:user
// @kthulu:generated:true
package user

import (
	"go.uber.org/fx"

	"vercel-test/internal/modules/user/api"
	"vercel-test/internal/modules/user/store"
	"vercel-test/internal/modules/user/core"
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
