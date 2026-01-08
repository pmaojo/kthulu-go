package adapterhttp

import "go.uber.org/fx"

// Module provides HTTP adapters for Fx.
var Module = fx.Options(
	fx.Provide(
		NewAuthHandler,
		NewUserHandler,
	),
)
