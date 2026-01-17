// @kthulu:module:debugmod
// @kthulu:generated:true
package debugmod

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"tournament-app/internal/modules/debugmod/api"
	"tournament-app/internal/modules/debugmod/store"
	"tournament-app/internal/modules/debugmod/core"
)

// Providers returns the Fx providers for the debugmod module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewDebugmodRepository,
                        core.NewDebugmodService,
                        api.NewDebugmodHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *api.DebugmodHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
