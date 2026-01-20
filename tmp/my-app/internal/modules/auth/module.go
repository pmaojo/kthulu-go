// @kthulu:module:auth
// @kthulu:generated:true
package auth

import (
	"go.uber.org/fx"

	"my-app/internal/modules/auth/api"
	"my-app/internal/modules/auth/store"
	"my-app/internal/modules/auth/core"
)

// Providers returns the Fx providers for the auth module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewAuthRepository,
                        core.NewAuthService,
                        api.NewAuthHandler,
                ),
        )
}
