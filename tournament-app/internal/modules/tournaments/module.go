// @kthulu:module:tournaments
// @kthulu:generated:true
package tournaments

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"tournament-app/internal/modules/tournaments/api"
	"tournament-app/internal/modules/tournaments/store"
	"tournament-app/internal/modules/tournaments/core"
)

// Providers returns the Fx providers for the tournaments module
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
