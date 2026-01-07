package adapters

import (
	"context"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/pkce"

	"github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/oauthsso/repository"
)

// FositeStorage adapts repositories to satisfy fosite.Storage.
type FositeStorage struct {
	clients  repository.ClientRepository
	sessions repository.SessionRepository
	tokens   repository.TokenRepository
}

var (
	_ fosite.Storage                = (*FositeStorage)(nil)
	_ oauth2.CoreStorage            = (*FositeStorage)(nil)
	_ oauth2.TokenRevocationStorage = (*FositeStorage)(nil)
	_ pkce.PKCERequestStorage       = (*FositeStorage)(nil)
)

// NewFositeStorage creates a new adapter instance.
func NewFositeStorage(c repository.ClientRepository, s repository.SessionRepository, t repository.TokenRepository) *FositeStorage {
	return &FositeStorage{clients: c, sessions: s, tokens: t}
}

// GetClient delegates to the underlying ClientRepository.
func (s *FositeStorage) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	return s.clients.GetClient(ctx, id)
}

// ClientAssertionJWTValid delegates to the ClientRepository.
func (s *FositeStorage) ClientAssertionJWTValid(ctx context.Context, jti string) error {
	return s.clients.ClientAssertionJWTValid(ctx, jti)
}

// SetClientAssertionJWT delegates to the ClientRepository.
func (s *FositeStorage) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error {
	return s.clients.SetClientAssertionJWT(ctx, jti, exp)
}

// CreateAuthorizeCodeSession persists the authorization code session via the SessionRepository.
func (s *FositeStorage) CreateAuthorizeCodeSession(ctx context.Context, code string, requester fosite.Requester) error {
	return s.sessions.CreateSession(ctx, code, requester)
}

// GetAuthorizeCodeSession retrieves an authorization code session from the SessionRepository.
func (s *FositeStorage) GetAuthorizeCodeSession(ctx context.Context, code string, session fosite.Session) (fosite.Requester, error) {
	req, err := s.sessions.GetSession(ctx, code, session)
	if err != nil {
		return nil, err
	}
	return attachSession(req, session), nil
}

// InvalidateAuthorizeCodeSession removes the session backing the authorization code.
func (s *FositeStorage) InvalidateAuthorizeCodeSession(ctx context.Context, code string) error {
	return s.sessions.DeleteSession(ctx, code)
}

// CreateAccessTokenSession persists an access token session via the TokenRepository.
func (s *FositeStorage) CreateAccessTokenSession(ctx context.Context, signature string, requester fosite.Requester) error {
	return s.tokens.CreateToken(ctx, signature, requester)
}

// GetAccessTokenSession retrieves an access token session and attaches the provided session state.
func (s *FositeStorage) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	req, err := s.tokens.GetToken(ctx, signature)
	if err != nil {
		return nil, err
	}
	return attachSession(req, session), nil
}

// DeleteAccessTokenSession removes an access token session via the TokenRepository.
func (s *FositeStorage) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	return s.tokens.DeleteToken(ctx, signature)
}

// CreateRefreshTokenSession persists a refresh token session via the TokenRepository.
func (s *FositeStorage) CreateRefreshTokenSession(ctx context.Context, signature string, accessSignature string, requester fosite.Requester) error {
	_ = accessSignature
	return s.tokens.CreateToken(ctx, signature, requester)
}

// GetRefreshTokenSession retrieves a refresh token session and hydrates it with the provided session state.
func (s *FositeStorage) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	req, err := s.tokens.GetToken(ctx, signature)
	if err != nil {
		return nil, err
	}
	return attachSession(req, session), nil
}

// DeleteRefreshTokenSession removes the refresh token session.
func (s *FositeStorage) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	return s.tokens.DeleteToken(ctx, signature)
}

// RotateRefreshToken revokes the refresh and associated access token sessions.
func (s *FositeStorage) RotateRefreshToken(ctx context.Context, requestID string, refreshTokenSignature string) error {
	_ = refreshTokenSignature
	if err := s.RevokeRefreshToken(ctx, requestID); err != nil {
		return err
	}
	return s.RevokeAccessToken(ctx, requestID)
}

// CreatePKCERequestSession stores the PKCE session via the SessionRepository.
func (s *FositeStorage) CreatePKCERequestSession(ctx context.Context, signature string, requester fosite.Requester) error {
	return s.sessions.CreateSession(ctx, signature, requester)
}

// GetPKCERequestSession fetches the PKCE session and hydrates it with the provided session state.
func (s *FositeStorage) GetPKCERequestSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	req, err := s.sessions.GetSession(ctx, signature, session)
	if err != nil {
		return nil, err
	}
	return attachSession(req, session), nil
}

// DeletePKCERequestSession removes the PKCE session from storage.
func (s *FositeStorage) DeletePKCERequestSession(ctx context.Context, signature string) error {
	return s.sessions.DeleteSession(ctx, signature)
}

// RevokeRefreshToken removes the refresh token identified by the request ID.
func (s *FositeStorage) RevokeRefreshToken(ctx context.Context, requestID string) error {
	return s.tokens.DeleteToken(ctx, requestID)
}

// RevokeAccessToken removes the access token identified by the request ID.
func (s *FositeStorage) RevokeAccessToken(ctx context.Context, requestID string) error {
	return s.tokens.DeleteToken(ctx, requestID)
}

func attachSession(request fosite.Requester, session fosite.Session) fosite.Requester {
	if request != nil && session != nil {
		request.SetSession(session)
	}
	return request
}
