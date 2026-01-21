// @kthulu:module:auth
// @kthulu:generated:true
package auth

import (
	"go.uber.org/fx"

	"feature-tour/internal/modules/auth/api"
	"feature-tour/internal/modules/auth/store"
	"feature-tour/internal/modules/auth/core"
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
