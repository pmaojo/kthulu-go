// @kthulu:module:invoice
// @kthulu:generated:true
package invoice

import (
	"go.uber.org/fx"

	"kitchen-sink-test/internal/modules/invoice/api"
	"kitchen-sink-test/internal/modules/invoice/store"
	"kitchen-sink-test/internal/modules/invoice/core"
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
