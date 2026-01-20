// @kthulu:module:calendar
// @kthulu:generated:true
package calendar

import (
	"go.uber.org/fx"

	"kitchen-sink-test/internal/modules/calendar/api"
	"kitchen-sink-test/internal/modules/calendar/store"
	"kitchen-sink-test/internal/modules/calendar/core"
)

// Providers returns the Fx providers for the calendar module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewCalendarRepository,
                        core.NewCalendarService,
                        api.NewCalendarHandler,
                ),
        )
}
