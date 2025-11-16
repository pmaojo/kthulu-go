package usecase

import (
	"context"
	"net/http"

	"github.com/ory/fosite"
)

// Token exchanges either an authorization code or a refresh token for a
// new access token (and potentially a new refresh token). The incoming
// request must contain the required OAuth2 parameters as defined by the
// respective grant type. The provided session will be populated with the
// data stored during the authorize step.
//
// The returned map contains the token response as defined by RFC 6749,
// including the access token, token type, optional refresh token and any
// additional parameters set by fosite.
func (uc *OAuthUseCase) Token(ctx context.Context, r *http.Request, session fosite.Session) (map[string]any, error) {
	provider := fosite.NewOAuth2Provider(uc.storage, &fosite.Config{})

	ar, err := provider.NewAccessRequest(ctx, r, session)
	if err != nil {
		return nil, err
	}

	resp, err := provider.NewAccessResponse(ctx, ar)
	if err != nil {
		return nil, err
	}

	return resp.ToMap(), nil
}
