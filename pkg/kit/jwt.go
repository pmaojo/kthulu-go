// @kthulu:core
package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager defines JWT signing and validation capabilities for both access and refresh tokens.
type TokenManager interface {
	SignAccessToken(claims jwt.Claims) (string, error)
	SignRefreshToken(claims jwt.Claims) (string, error)
	ValidateAccessToken(token string) (jwt.MapClaims, error)
	ValidateRefreshToken(token string) (jwt.MapClaims, error)
	GetAccessTokenTTL() time.Duration
	GetRefreshTokenTTL() time.Duration
}

type jwtManager struct {
	accessSecret    []byte
	refreshSecret   []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewJWT constructs a JWT token manager using the application's JWT configuration.
func NewJWT(cfg *Config) TokenManager {
	return &jwtManager{
		accessSecret:    []byte(cfg.JWT.Secret),
		refreshSecret:   []byte(cfg.JWT.RefreshSecret),
		accessTokenTTL:  cfg.JWT.AccessTokenTTL,
		refreshTokenTTL: cfg.JWT.RefreshTokenTTL,
	}
}

// SignAccessToken creates a signed JWT access token string for the provided claims using HS256.
func (j *jwtManager) SignAccessToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.accessSecret)
}

// SignRefreshToken creates a signed JWT refresh token string for the provided claims using HS256.
func (j *jwtManager) SignRefreshToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.refreshSecret)
}

// ValidateAccessToken parses and validates an access token string, returning its MapClaims.
func (j *jwtManager) ValidateAccessToken(tokenStr string) (jwt.MapClaims, error) {
	return j.validateToken(tokenStr, j.accessSecret)
}

// ValidateRefreshToken parses and validates a refresh token string, returning its MapClaims.
func (j *jwtManager) ValidateRefreshToken(tokenStr string) (jwt.MapClaims, error) {
	return j.validateToken(tokenStr, j.refreshSecret)
}

// GetAccessTokenTTL returns the configured access token time-to-live duration.
func (j *jwtManager) GetAccessTokenTTL() time.Duration {
	return j.accessTokenTTL
}

// GetRefreshTokenTTL returns the configured refresh token time-to-live duration.
func (j *jwtManager) GetRefreshTokenTTL() time.Duration {
	return j.refreshTokenTTL
}

// validateToken is a helper method to validate tokens with the provided secret.
func (j *jwtManager) validateToken(tokenStr string, secret []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
