// @kthulu:module:mail
// @kthulu:generated:true
package mail

import (
	"go.uber.org/fx"

	"feature-tour/internal/modules/mail/api"
	"feature-tour/internal/modules/mail/store"
	"feature-tour/internal/modules/mail/core"
)

// Providers returns the Fx providers for the mail module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewMailRepository,
                        core.NewMailService,
                        api.NewMailHandler,
                ),
        )
}
