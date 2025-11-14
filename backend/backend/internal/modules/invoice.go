// @kthulu:module:invoices
package modules

import (
	"go.uber.org/fx"

	adapterhttp "backend/internal/adapters/http"
	"backend/internal/usecase"
)

// InvoiceModule provides invoice functionality.
// Repositories are provided by SharedRepositoryProviders to avoid duplication.
var InvoiceModule = fx.Options(
	// Use cases
	fx.Provide(
		usecase.NewInvoiceUseCase,
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewInvoiceHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.InvoiceHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
