// @kthulu:module:user
package modules

import (
	"go.uber.org/fx"

	adapterhttp "backend/internal/adapters/http"
	"backend/internal/usecase"
)

// UserModule provides user profile functionality.
// Repositories are provided by SharedRepositoryProviders to avoid duplication.
var UserModule = fx.Options(
	// Use cases
	fx.Provide(
		usecase.NewUserUseCase,
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewUserHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.UserHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
