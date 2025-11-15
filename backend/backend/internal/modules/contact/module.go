// @kthulu:module:contact
package contact

import (
	"go.uber.org/fx"

	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// Module provides fx.Options for contact (CRM) module.
// Includes contact/customer management.
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			db.NewContactRepository,
			fx.As(new(repository.ContactRepository)),
		),
	),
)
