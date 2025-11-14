// @kthulu:module:calendar
package modules

import (
	"go.uber.org/fx"

	adapterhttp "backend/internal/adapters/http"
	"backend/internal/usecase"
)

// CalendarModule provides calendar and appointment scheduling functionality
var CalendarModule = fx.Options(
	// Use cases
	fx.Provide(
		usecase.NewCalendarUseCase,
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewCalendarHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.CalendarHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
