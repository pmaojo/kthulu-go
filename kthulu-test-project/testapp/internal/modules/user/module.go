// @kthulu:module:user
// @kthulu:generated:true
package user

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"testapp/internal/modules/user/api"
	"testapp/internal/modules/user/store"
	"testapp/internal/modules/user/core"
)

// Providers returns the Fx providers for the user module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewUserRepository,
                        core.NewUserService,
                        api.NewUserHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *api.UserHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
