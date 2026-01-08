// @kthulu:module:auth
// @kthulu:generated:true
package auth

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"airbnb-v2/internal/modules/auth/handlers"
	"airbnb-v2/internal/modules/auth/repository"
	"airbnb-v2/internal/modules/auth/service"
)

// Providers returns the Fx providers for the auth module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewAuthRepository,
                        service.NewAuthService,
                        handlers.NewAuthHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.AuthHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
