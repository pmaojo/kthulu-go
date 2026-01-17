// @kthulu:module:auth
// @kthulu:generated:true
package auth

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"tournament-app/internal/modules/auth/api"
	"tournament-app/internal/modules/auth/store"
	"tournament-app/internal/modules/auth/core"
)

// Providers returns the Fx providers for the auth module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewAuthRepository,
                        core.NewAuthService,
                        api.NewAuthHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *api.AuthHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
