// @kthulu:module:auth
package modules

import (
	"go.uber.org/fx"

	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

// AuthModule provides authentication functionality.
// Repositories are injected via the ModuleSet provider map to avoid duplication.
var AuthModule = fx.Options(
	// Use cases
	fx.Provide(
		usecase.NewAuthUseCase,
		usecase.NewAuthService,
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewAuthHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.AuthHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
