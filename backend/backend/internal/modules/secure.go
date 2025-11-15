package modules

import (
	"go.uber.org/fx"

	"github.com/pmaojo/kthulu-go/backend/internal/secure"
)

// SecureModule exposes security utilities via HTTP.
var SecureModule = fx.Options(
	fx.Provide(
		secure.NewHandler,
	),
	fx.Invoke(func(h *secure.Handler, registry *RouteRegistry) {
		registry.Register(h)
	}),
)
