// @kthulu:module:participants
// @kthulu:generated:true
package participants

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"tournament-app/internal/modules/participants/api"
	"tournament-app/internal/modules/participants/store"
	"tournament-app/internal/modules/participants/core"
)

// Providers returns the Fx providers for the participants module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewParticipantRepository,
                        core.NewParticipantService,
                        api.NewParticipantHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *api.ParticipantHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
