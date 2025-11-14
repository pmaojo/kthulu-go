// @kthulu:module:inventory
package modules

import (
	"go.uber.org/fx"

	adapterhttp "backend/internal/adapters/http"
	"backend/internal/usecase"
)

// InventoryModule provides inventory management functionality
var InventoryModule = fx.Options(
	// Use cases
	fx.Provide(
		usecase.NewInventoryUseCase,
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewInventoryHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.InventoryHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
