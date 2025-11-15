// @kthulu:module:templates
package modules

import (
	"go.uber.org/fx"

	adapterhttp "github.com/kthulu/kthulu-go/backend/internal/adapters/http"
	"github.com/kthulu/kthulu-go/backend/internal/infrastructure/db"
	"github.com/kthulu/kthulu-go/backend/internal/usecase"
)

// TemplatesModule provides template management functionality.
var TemplatesModule = fx.Options(
	// Repositories
	fx.Provide(
		db.NewTemplateRepository,
		db.NewTemplateRegistryRepository,
	),

	// Use cases
	fx.Provide(
		usecase.NewTemplateUseCase,
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewTemplateHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.TemplateHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
