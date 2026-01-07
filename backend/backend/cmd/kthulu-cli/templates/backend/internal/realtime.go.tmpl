// @kthulu:module:realtime
package modules

import (
	"go.uber.org/fx"

	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
	adapterrealtime "github.com/pmaojo/kthulu-go/backend/internal/adapters/realtime"
	"github.com/pmaojo/kthulu-go/backend/internal/domain/repository"
	usecasert "github.com/pmaojo/kthulu-go/backend/internal/usecase/realtime"
)

// RealtimeModule wires realtime capabilities.
var RealtimeModule = fx.Options(
	adapterrealtime.Module,
	fx.Provide(
		repository.NewInMemoryConnectionRepository,
		usecasert.NewService,
		adapterhttp.NewRealtimeHandler,
	),
	fx.Invoke(func(h *adapterhttp.RealtimeHandler, rr *RouteRegistry) {
		rr.Register(h)
	}),
)
