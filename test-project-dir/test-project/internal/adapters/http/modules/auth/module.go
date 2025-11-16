// @kthulu:module:auth
// @kthulu:category:Custom
package auth

import "go.uber.org/fx"

// Providers returns the Fx providers for the auth module
func Providers() fx.Option {
	return fx.Options(
		fx.Provide(
			NewAuthRepository,
			NewAuthService,
			NewAuthHandler,
		),
	)
}
