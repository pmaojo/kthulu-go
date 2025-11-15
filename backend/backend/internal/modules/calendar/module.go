// @kthulu:module:calendar
package calendar

import (
	"go.uber.org/fx"

	"github.com/kthulu/kthulu-go/backend/internal/infrastructure/db"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
)

// Module provides fx.Options for calendar (scheduling) module.
// Includes calendar and appointment management.
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			db.NewCalendarRepository,
			fx.As(new(repository.CalendarRepository)),
		),
	),
)
