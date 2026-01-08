package domain

import (
	"errors"
	"time"
)

// Token represents an issued token.
type Token struct {
	Claims    map[string]any
	JTI       string
	ExpiresAt time.Time
}

// NewToken creates a new Token with minimal validation.
func NewToken(claims map[string]any, jti string, expiresAt time.Time) (*Token, error) {
	if jti == "" {
		return nil, errors.New("jti is required")
	}
	if expiresAt.IsZero() {
		return nil, errors.New("expiresAt is required")
	}
	if claims == nil {
		claims = make(map[string]any)
	}
	return &Token{Claims: claims, JTI: jti, ExpiresAt: expiresAt}, nil
}
