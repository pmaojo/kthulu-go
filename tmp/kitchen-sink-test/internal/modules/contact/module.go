// @kthulu:module:contact
// @kthulu:generated:true
package contact

import (
	"go.uber.org/fx"

	"kitchen-sink-test/internal/modules/contact/api"
	"kitchen-sink-test/internal/modules/contact/store"
	"kitchen-sink-test/internal/modules/contact/core"
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
