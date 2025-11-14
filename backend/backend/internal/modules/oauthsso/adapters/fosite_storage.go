package adapters

import (
	"context"
	"time"

	"github.com/ory/fosite"

	"backend/internal/modules/oauthsso/repository"
)

// FositeStorage adapts repositories to satisfy fosite.Storage.
type FositeStorage struct {
	clients  repository.ClientRepository
	sessions repository.SessionRepository
	tokens   repository.TokenRepository
}

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
