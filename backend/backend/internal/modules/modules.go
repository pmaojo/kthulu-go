// @kthulu:module:modules
package modules

import (
	"go.uber.org/fx"

	adapterhttp "backend/internal/adapters/http"
	"backend/internal/infrastructure/db"
	"backend/internal/usecase"
)

// ModulesModule provides module catalog functionality.
var ModulesModule = fx.Options(
	// Repository
	fx.Provide(
		db.NewModuleRepository,
	),

	// Use cases
	fx.Provide(
		usecase.NewModuleUseCase,
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewModuleHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.ModuleHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
