// @kthulu:core
package modules

import (
	"go.uber.org/fx"

	adapterhttp "backend/internal/adapters/http"
)

// StaticModule provides static file serving functionality for the frontend.
// This module enables serving the built frontend application from the same binary.
var StaticModule = fx.Options(
	// HTTP handlers
	fx.Provide(
		adapterhttp.NewStaticHandler,
	),

	// Register routes - this should be registered last to avoid conflicts
	fx.Invoke(func(handler *adapterhttp.StaticHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
