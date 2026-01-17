// @kthulu:module:participant
// @kthulu:generated:true
package participant

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"tournament-api/internal/modules/participant/api"
	"tournament-api/internal/modules/participant/store"
	"tournament-api/internal/modules/participant/core"
)

// Providers returns the Fx providers for the participant module
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
