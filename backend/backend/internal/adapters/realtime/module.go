package adapterrealtime

import "go.uber.org/fx"

// Module provides the realtime adapter for Fx.
var Module = fx.Options(
	fx.Provide(NewHandler),
)
