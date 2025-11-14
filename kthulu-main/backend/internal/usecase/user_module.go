// @kthulu:module:user
package usecase

import "go.uber.org/fx"

// UserModule provides user-related use cases for Fx.
var UserModule = fx.Options(
	fx.Provide(NewUserUseCase),
)
