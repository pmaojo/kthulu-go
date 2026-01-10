// @kthulu:module:reviews
// @kthulu:generated:true
package reviews

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"test-app/internal/modules/reviews/handlers"
	"test-app/internal/modules/reviews/repository"
	"test-app/internal/modules/reviews/service"
)

// Providers returns the Fx providers for the reviews module
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
