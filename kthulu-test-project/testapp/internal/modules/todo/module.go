// @kthulu:module:todo
// @kthulu:generated:true
package todo

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"testapp/internal/modules/todo/api"
	"testapp/internal/modules/todo/store"
	"testapp/internal/modules/todo/core"
)

// Providers returns the Fx providers for the todo module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewTodoRepository,
                        core.NewTodoService,
                        api.NewTodoHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *api.TodoHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
