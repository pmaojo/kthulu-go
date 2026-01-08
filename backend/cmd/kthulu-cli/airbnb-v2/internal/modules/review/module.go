// @kthulu:module:review
// @kthulu:generated:true
package review

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"airbnb-v2/internal/modules/review/handlers"
	"airbnb-v2/internal/modules/review/repository"
	"airbnb-v2/internal/modules/review/service"
)

// Providers returns the Fx providers for the review module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.NewReviewRepository,
                        service.NewReviewService,
                        handlers.NewReviewHandler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.ReviewHandler) {
                        h.RegisterRoutes(r)
                }),
        )
}
