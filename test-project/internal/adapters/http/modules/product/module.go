// @kthulu:module:product
// @kthulu:category:Custom
package product

import "go.uber.org/fx"

// Providers returns the Fx providers for the product module
func Providers() fx.Option {
	return fx.Options(
		fx.Provide(
			NewProductRepository,
			NewProductService,
			NewProductHandler,
		),
	)
}
