// @kthulu:module:auth
// @kthulu:generated:true
package auth

import (
	"go.uber.org/fx"

	"kitchen-sink-test/internal/modules/auth/api"
	"kitchen-sink-test/internal/modules/auth/store"
	"kitchen-sink-test/internal/modules/auth/core"
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
