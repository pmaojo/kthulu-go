// @kthulu:module:scheduler
// @kthulu:generated:true
package scheduler

import (
	"go.uber.org/fx"

	"feature-tour/internal/modules/scheduler/api"
	"feature-tour/internal/modules/scheduler/store"
	"feature-tour/internal/modules/scheduler/core"
)

// Providers returns the Fx providers for the scheduler module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewSchedulerRepository,
                        core.NewSchedulerService,
                        api.NewSchedulerHandler,
                ),
        )
}
