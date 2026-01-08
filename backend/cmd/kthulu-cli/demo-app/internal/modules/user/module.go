// @kthulu:module:user
// @kthulu:generated:true
package user

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"demo-app/internal/modules/user/handlers"
	"demo-app/internal/modules/user/repository"
	"demo-app/internal/modules/user/service"
)

// Providers returns the Fx providers for the user module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewUserRepository,
                        service.NewUserService,
                        handlers.NewUserHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.UserHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
