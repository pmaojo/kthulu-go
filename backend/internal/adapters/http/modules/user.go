// @kthulu:module:user
package modules

import (
	"go.uber.org/fx"

	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

// UserModule provides user profile functionality.
// Repositories are injected via the ModuleSet provider map to avoid duplication.
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
