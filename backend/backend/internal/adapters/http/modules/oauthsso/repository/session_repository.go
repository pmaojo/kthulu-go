package repository

import (
	"context"

	"github.com/ory/fosite"
)

// SessionRepository defines persistence behavior for OAuth sessions.
type SessionRepository interface {
	// CreateSession stores the request under the given signature.
	CreateSession(ctx context.Context, signature string, requester fosite.Requester) error

	// GetSession retrieves a previously stored session.
	GetSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error)

	// DeleteSession removes the session associated with the signature.
	DeleteSession(ctx context.Context, signature string) error
}
