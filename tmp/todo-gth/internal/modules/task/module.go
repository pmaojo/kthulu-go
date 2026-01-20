// @kthulu:module:task
// @kthulu:generated:true
package task

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"todo-gth/internal/modules/task/api"
	"todo-gth/internal/modules/task/store"
	"todo-gth/internal/modules/task/core"
)

// Providers returns the Fx providers for the task module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewTaskRepository,
                        core.NewTaskService,
                        api.NewTaskHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *api.TaskHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
