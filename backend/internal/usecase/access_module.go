// @kthulu:module:access
package usecase

import "go.uber.org/fx"

// AccessModule provides access control use cases for Fx.
var AccessModule = fx.Options(
	fx.Provide(NewAccessUseCase),
)
