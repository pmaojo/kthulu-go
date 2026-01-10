package observability

import "go.uber.org/fx"

// Module provides observability dependencies.
var Module = fx.Options(
	fx.Provide(
		NewLogger,
		GetZapLogger,
		NewTracerProvider,
		NewMetricsProvider,
	),
)
