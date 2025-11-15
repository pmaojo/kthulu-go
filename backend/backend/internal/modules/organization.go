// @kthulu:module:org
package modules

import (
	"go.uber.org/fx"

	adapterhttp "github.com/kthulu/kthulu-go/backend/internal/adapters/http"
	"github.com/kthulu/kthulu-go/backend/internal/usecase"
)

// OrganizationModule provides organization functionality.
// Repositories are provided by SharedRepositoryProviders to avoid duplication.
var OrganizationModule = fx.Options(
	// Use cases
	fx.Provide(
		usecase.NewOrganizationUseCase,
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewOrganizationHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.OrganizationHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
