package usecase

import (
	"context"
	"net/http"

	"github.com/ory/fosite"
)

// Authorize handles the OAuth2 authorization code flow with PKCE.
// It assumes that the caller has already authenticated the end user
// and obtained consent for the requested scopes. The provided session
// should contain any subject or additional claims that should be
// persisted for the authorization code exchange.
//
// The method returns the generated authorization code which can later
// be exchanged at the token endpoint.
func (uc *OAuthUseCase) Authorize(ctx context.Context, r *http.Request, session fosite.Session) (string, error) {
	// Create a new OAuth2 provider using the configured storage.
	provider := fosite.NewOAuth2Provider(uc.storage, &fosite.Config{})

	// Parse the incoming authorization request.
	ar, err := provider.NewAuthorizeRequest(ctx, r)
	if err != nil {
		return "", err
	}

	// Grant all requested scopes. In a real world scenario the
	// application would present a consent screen and selectively grant
	// scopes. For this exercise we simply assume consent.
	for _, scope := range ar.GetRequestedScopes() {
		ar.GrantScope(scope)
	}

	// Create the authorization response which issues the authorization
	// code. PKCE validation is handled internally by fosite if the
	// request contains the appropriate verifier parameters.
	resp, err := provider.NewAuthorizeResponse(ctx, ar, session)
	if err != nil {
		return "", err
	}

	// The authorization code is available on the response structure.
	return resp.GetCode(), nil
}
