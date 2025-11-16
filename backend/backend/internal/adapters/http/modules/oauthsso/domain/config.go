package domain

import (
	"fmt"
	"os"
	"time"
)

// Config holds configuration for OAuth SSO module.
type Config struct {
	Issuer               string
	JWTSecret            string
	AccessTokenLifetime  time.Duration
	RefreshTokenLifetime time.Duration
}

// NewConfigFromEnv loads configuration from environment variables.
func NewConfigFromEnv() (*Config, error) {
	issuer := os.Getenv("SSO_ISSUER")
	if issuer == "" {
		return nil, fmt.Errorf("SSO_ISSUER is required")
	}

	secret := os.Getenv("SSO_JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("SSO_JWT_SECRET is required")
	}

	accessTTL, err := time.ParseDuration(getEnvWithDefault("SSO_ACCESS_TOKEN_LIFETIME", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid SSO_ACCESS_TOKEN_LIFETIME: %w", err)
	}

	refreshTTL, err := time.ParseDuration(getEnvWithDefault("SSO_REFRESH_TOKEN_LIFETIME", "7d"))
	if err != nil {
		return nil, fmt.Errorf("invalid SSO_REFRESH_TOKEN_LIFETIME: %w", err)
	}

	return &Config{
		Issuer:               issuer,
		JWTSecret:            secret,
		AccessTokenLifetime:  accessTTL,
		RefreshTokenLifetime: refreshTTL,
	}, nil
}

func getEnvWithDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
