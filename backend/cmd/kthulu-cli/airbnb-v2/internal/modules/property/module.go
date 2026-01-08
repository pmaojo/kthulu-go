// @kthulu:module:property
// @kthulu:generated:true
package property

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"airbnb-v2/internal/modules/property/handlers"
	"airbnb-v2/internal/modules/property/repository"
	"airbnb-v2/internal/modules/property/service"
)

// Providers returns the Fx providers for the property module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewPropertyRepository,
                        service.NewPropertyService,
                        handlers.NewPropertyHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.PropertyHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
