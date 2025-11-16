// @kthulu:module:user
// @kthulu:category:Custom
package user

import "go.uber.org/fx"

// Providers returns the Fx providers for the user module
func Providers() fx.Option {
	return fx.Options(
		fx.Provide(
			NewUserRepository,
			NewUserService,
			NewUserHandler,
		),
	)
}
