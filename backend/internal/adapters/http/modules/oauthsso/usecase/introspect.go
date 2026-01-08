package usecase

import (
	"context"

	"github.com/ory/fosite"
)

// Introspect validates an OAuth2 access or refresh token. It returns the
// token use (access or refresh) together with the access request
// associated to the token if validation succeeds. External resource
// servers can use the returned access request to obtain information about
// the client or granted scopes.
func (uc *OAuthUseCase) Introspect(ctx context.Context, token string, session fosite.Session, scopes ...string) (fosite.TokenUse, fosite.AccessRequester, error) {
	provider := fosite.NewOAuth2Provider(uc.storage, &fosite.Config{})

	tokenUse, ar, err := provider.IntrospectToken(ctx, token, fosite.AccessToken, session, scopes...)
	if err != nil {
		return "", nil, err
	}
	return tokenUse, ar, nil
}
