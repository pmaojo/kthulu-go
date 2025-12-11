// @kthulu:module:cli
package modules

import (
	"go.uber.org/fx"

	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
)

// CliModule provides CLI execution endpoints.
var CliModule = fx.Options(
	// HTTP handlers
	fx.Provide(
		adapterhttp.NewCliHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.CliHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
