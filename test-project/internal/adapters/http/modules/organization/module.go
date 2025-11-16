// @kthulu:module:organization
// @kthulu:category:Custom
package organization

import "go.uber.org/fx"

// Providers returns the Fx providers for the organization module
func Providers() fx.Option {
	return fx.Options(
		fx.Provide(
			NewOrganizationRepository,
			NewOrganizationService,
			NewOrganizationHandler,
		),
	)
}
