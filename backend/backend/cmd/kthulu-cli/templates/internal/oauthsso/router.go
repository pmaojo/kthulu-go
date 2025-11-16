package oauthsso

import (
	"github.com/go-chi/chi/v5"
	"github.com/ory/fosite"

	oauthhttp "github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/oauthsso/adapters/http"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/oauthsso/usecase"
)

// Router wires OAuth2 handlers under the /oauth prefix.
type Router struct {
	handlers *oauthhttp.Handlers
}

// NewRouter constructs a Router with required dependencies.
func NewRouter(provider fosite.OAuth2Provider, uc *usecase.OAuthUseCase) *Router {
	return &Router{handlers: oauthhttp.NewHandlers(provider, uc)}
}

// RegisterRoutes mounts the OAuth2 routes on the given router.
func (r *Router) RegisterRoutes(root chi.Router) {
	root.Route("/oauth", func(rr chi.Router) {
		rr.Get("/authorize", r.handlers.Authorize)
		rr.Post("/token", r.handlers.Token)
		rr.Get("/userinfo", r.handlers.UserInfo)
		rr.Get("/jwks.json", r.handlers.JWKS)
		rr.Post("/introspect", r.handlers.Introspect)
		rr.Post("/revoke", r.handlers.Revoke)
	})
}
