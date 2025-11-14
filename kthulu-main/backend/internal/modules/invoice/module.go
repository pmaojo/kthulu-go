// @kthulu:module:invoice
package invoice

import (
	"go.uber.org/fx"

	"backend/internal/infrastructure/db"
	"backend/internal/repository"
)

// Module provides fx.Options for invoice (billing) module.
// Includes invoice and billing management.
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			db.NewInvoiceRepository,
			fx.As(new(repository.InvoiceRepository)),
		),
	),
)
