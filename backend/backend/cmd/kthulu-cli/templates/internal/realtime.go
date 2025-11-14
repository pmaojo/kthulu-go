// @kthulu:module:realtime
package modules

import (
	"go.uber.org/fx"

	adapterhttp "backend/internal/adapters/http"
	adapterrealtime "backend/internal/adapters/realtime"
	"backend/internal/repository"
	usecasert "backend/internal/usecase/realtime"
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
