package repository

import (
	"context"
	"time"

	"github.com/ory/fosite"
)

// ClientRepository defines persistence behavior for OAuth clients.
type ClientRepository interface {
	// GetClient retrieves a client by its identifier.
	GetClient(ctx context.Context, id string) (fosite.Client, error)

	// ClientAssertionJWTValid checks if a JTI has been used previously.
	ClientAssertionJWTValid(ctx context.Context, jti string) error

	// SetClientAssertionJWT marks a JTI as used until the given expiry time.
	SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error
}
