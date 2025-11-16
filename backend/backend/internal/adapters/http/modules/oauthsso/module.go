// @kthulu:module:oauthsso
package oauthsso

import (
	"go.uber.org/fx"

	oauthdb "github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/oauthsso/infrastructure/db"
	oauthrepo "github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/oauthsso/repository"
)

// Module provides fx.Options for OAuth SSO (external authentication) module.
// Includes OAuth client, session, and token management.
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			oauthdb.NewClientRepository,
			fx.As(new(oauthrepo.ClientRepository)),
		),
		fx.Annotate(
			oauthdb.NewSessionRepository,
			fx.As(new(oauthrepo.SessionRepository)),
		),
		fx.Annotate(
			oauthdb.NewTokenRepository,
			fx.As(new(oauthrepo.TokenRepository)),
		),
	),
)
