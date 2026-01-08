// @kthulu:module:auth
package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

// RefreshToken-related errors
var (
	ErrTokenExpired    = errors.New("refresh token expired")
	ErrInvalidToken    = errors.New("invalid refresh token")
	ErrTokenNotFound   = errors.New("refresh token not found")
	ErrTokenGeneration = errors.New("failed to generate token")
)

// RefreshToken represents a JWT refresh token
type RefreshToken struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	User      *User     `json:"user,omitempty"`
}

// NewRefreshToken creates a new refresh token for a user. It returns the hashed
// token stored in the RefreshToken struct and the raw token value which should be
// used as the JWT ID (jti).
func NewRefreshToken(userID uint, ttl time.Duration) (*RefreshToken, string, error) {
	if userID == 0 {
		return nil, "", errors.New("user ID is required")
	}

	if ttl <= 0 {
		return nil, "", errors.New("TTL must be positive")
	}

	rawToken, err := generateSecureToken(32) // 32 bytes = 64 hex characters
	if err != nil {
		return nil, "", ErrTokenGeneration
	}

	hash := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(hash[:])

	now := time.Now()
	refreshToken := &RefreshToken{
		UserID:    userID,
		Token:     hashedToken,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}

	return refreshToken, rawToken, nil
}

// IsExpired returns true if the token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid returns true if the token is valid (not expired)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired()
}

// TimeUntilExpiry returns the duration until the token expires
func (rt *RefreshToken) TimeUntilExpiry() time.Duration {
	if rt.IsExpired() {
		return 0
	}
	return rt.ExpiresAt.Sub(time.Now())
}

// Extend extends the token expiry by the given duration
func (rt *RefreshToken) Extend(duration time.Duration) {
	rt.ExpiresAt = rt.ExpiresAt.Add(duration)
}

// Revoke marks the token as expired (effectively revoking it)
func (rt *RefreshToken) Revoke() {
	rt.ExpiresAt = time.Now().Add(-time.Hour) // Set to past time
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// TokenInfo provides information about the token without exposing the token value
type TokenInfo struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	IsExpired bool      `json:"isExpired"`
}

// GetInfo returns token information without the sensitive token value
func (rt *RefreshToken) GetInfo() TokenInfo {
	return TokenInfo{
		ID:        rt.ID,
		UserID:    rt.UserID,
		ExpiresAt: rt.ExpiresAt,
		CreatedAt: rt.CreatedAt,
		IsExpired: rt.IsExpired(),
	}
}
