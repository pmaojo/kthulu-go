package adapterrealtime

// @kthulu:module:realtime

import "go.uber.org/fx"

// Module provides the realtime adapter for Fx.
var Module = fx.Options(
	fx.Provide(NewHandler),
)
