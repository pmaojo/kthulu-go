package tournamentv2

import (
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"tournament-api/internal/modules/tournamentv2/api"
	"tournament-api/internal/modules/tournamentv2/core"
	"tournament-api/internal/modules/tournamentv2/store"
)

type Module struct {
	Handler *api.TournamentV2Handler
	Service core.TournamentV2Service
	Store   core.TournamentV2Repository
}

func NewModule(db *gorm.DB) *Module {
	s := store.NewTournamentV2Store(db)
	svc := core.NewTournamentV2Service(s)
	h := api.NewTournamentV2Handler(svc)

	return &Module{
		Handler: h,
		Service: svc,
		Store:   s,
	}
}

func (m *Module) RegisterRoutes(router *mux.Router) {
	m.Handler.RegisterRoutes(router)
}
