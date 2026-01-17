// @kthulu:module:tournament
// @kthulu:generated:true
package tournament

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"tournament-api/internal/modules/tournament/api"
	"tournament-api/internal/modules/tournament/store"
	"tournament-api/internal/modules/tournament/core"
)

// Providers returns the Fx providers for the tournament module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewTournamentRepository,
                        core.NewTournamentService,
                        api.NewTournamentHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *api.TournamentHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
