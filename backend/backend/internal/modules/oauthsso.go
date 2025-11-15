// @kthulu:module:oauth-sso
package modules

import (
	modulesoauth "github.com/kthulu/kthulu-go/backend/internal/modules/oauthsso"
	oauthadapters "github.com/kthulu/kthulu-go/backend/internal/modules/oauthsso/adapters"
	"github.com/kthulu/kthulu-go/backend/internal/modules/oauthsso/domain"
	"github.com/kthulu/kthulu-go/backend/internal/modules/oauthsso/repository"
	"github.com/kthulu/kthulu-go/backend/internal/modules/oauthsso/usecase"

	"github.com/ory/fosite"
	"go.uber.org/fx"
)

// OAuthSSOModule exports OAuth SSO dependencies.
var OAuthSSOModule = fx.Options(
	fx.Provide(
		domain.NewConfigFromEnv,
		provideFositeStorage,
		provideOAuthUseCases,
		provideRouter,
	),
	fx.Invoke(registerRoutes),
)

func provideFositeStorage(
	clients repository.ClientRepository,
	sessions repository.SessionRepository,
	tokens repository.TokenRepository,
) fosite.Storage {
	return oauthadapters.NewFositeStorage(clients, sessions, tokens)
}

func provideOAuthUseCases(cfg *domain.Config, storage fosite.Storage) *usecase.OAuthUseCase {
	return usecase.NewOAuthUseCase(cfg, storage)
}

func provideRouter(storage fosite.Storage, uc *usecase.OAuthUseCase) *modulesoauth.Router {
	provider := fosite.NewOAuth2Provider(storage, &fosite.Config{})
	return modulesoauth.NewRouter(provider, uc)
}

func registerRoutes(rr *RouteRegistry, r *modulesoauth.Router) {
	rr.Register(r)
}
