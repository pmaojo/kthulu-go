// @kthulu:module:invoice
// @kthulu:generated:true
package invoice

import (
	"go.uber.org/fx"

	"feature-tour/internal/modules/invoice/api"
	"feature-tour/internal/modules/invoice/store"
	"feature-tour/internal/modules/invoice/core"
)

// Providers returns the Fx providers for the invoice module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewInvoiceRepository,
                        core.NewInvoiceService,
                        api.NewInvoiceHandler,
                ),
        )
}
