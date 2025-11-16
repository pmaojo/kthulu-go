// @kthulu:module:projects
package modules

import (
	"go.uber.org/fx"

	adapterhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http"
	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

// ProjectsModule provides project management functionality.
var ProjectsModule = fx.Options(
	// Repository
	fx.Provide(
		db.NewProjectRepository,
	),

	// Use cases
	fx.Provide(
		usecase.NewProjectUseCase,
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewProjectHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.ProjectHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
