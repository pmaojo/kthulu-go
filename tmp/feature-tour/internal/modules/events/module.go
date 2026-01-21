// @kthulu:module:events
// @kthulu:generated:true
package events

import (
	"go.uber.org/fx"

	"feature-tour/internal/modules/events/api"
	"feature-tour/internal/modules/events/store"
	"feature-tour/internal/modules/events/core"
)

// Providers returns the Fx providers for the events module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewEventRepository,
                        core.NewEventService,
                        api.NewEventHandler,
                ),
        )
}
