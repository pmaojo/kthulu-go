// @kthulu:module:modules
package modules

import (
	"go.uber.org/fx"

	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
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
