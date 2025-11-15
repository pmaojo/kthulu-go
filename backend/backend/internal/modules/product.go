// @kthulu:module:products
package modules

import (
	"go.uber.org/fx"

	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

// ProductModule provides product functionality.
// Repositories are injected via the ModuleSet provider map to avoid duplication.
var ProductModule = fx.Options(
	// Use cases
	fx.Provide(
		usecase.NewProductUseCase,
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewProductHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.ProductHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
