// @kthulu:module:matches
// @kthulu:generated:true
package matches

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"tournament-app/internal/modules/matches/api"
	"tournament-app/internal/modules/matches/store"
	"tournament-app/internal/modules/matches/core"
)

// Providers returns the Fx providers for the matches module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewMatchRepository,
                        core.NewMatchService,
                        api.NewMatchHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *api.MatchHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
