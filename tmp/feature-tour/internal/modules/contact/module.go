// @kthulu:module:contact
// @kthulu:generated:true
package contact

import (
	"go.uber.org/fx"

	"feature-tour/internal/modules/contact/api"
	"feature-tour/internal/modules/contact/store"
	"feature-tour/internal/modules/contact/core"
)

// Providers returns the Fx providers for the contact module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewContactRepository,
                        core.NewContactService,
                        api.NewContactHandler,
                ),
        )
}
