// @kthulu:module:auth
package auth

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"
)

func Providers() fx.Option {
	return fx.Options(
		fx.Provide(
			NewAuthRepository,
			NewAuthService,
			NewAuthHandler,
		),
		fx.Invoke(func(r *mux.Router, h *AuthHandler) {
			h.RegisterRoutes(r)
		}),
	)
}
