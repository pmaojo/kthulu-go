// @kthulu:core
package modules

import (
	"go.uber.org/fx"

	adapterhttp "github.com/kthulu/kthulu-go/backend/internal/adapters/http"
)

// HealthModule provides health check functionality
var HealthModule = fx.Options(
	// HTTP handlers
	fx.Provide(
		adapterhttp.NewHealthHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.HealthHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
