// @kthulu:module:auth
package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"backend/core"
	"backend/internal/domain"
	"backend/internal/repository"
)

// AuthService provides authentication services with improved token management
type AuthService struct {
	users         repository.UserRepository
	refreshTokens repository.RefreshTokenRepository
	roles         repository.RoleRepository
	tokenStorage  repository.TokenStorage
	tokens        core.TokenManager
	notifier      repository.NotificationProvider
	logger        core.Logger
	authUseCase   *AuthUseCase
}

// NewAuthService creates a new AuthService with improved token management
func NewAuthService(
	users repository.UserRepository,
	refreshTokens repository.RefreshTokenRepository,
	roles repository.RoleRepository,
	tokenStorage repository.TokenStorage,
	tokenManager core.TokenManager,
	notifier repository.NotificationProvider,
	logger core.Logger,
) *AuthService {
	authUC := &AuthUseCase{
		users:         users,
		refreshTokens: refreshTokens,
		roles:         roles,
		tokens:        tokenManager,
		notifier:      notifier,
		logger:        logger,
	}

	return &AuthService{
		users:         users,
		refreshTokens: refreshTokens,
		roles:         roles,
		tokenStorage:  tokenStorage,
		tokens:        tokenManager,
		notifier:      notifier,
		logger:        logger,
		authUseCase:   authUC,
	}
}

// TokenProvider defines a function that provides tokens
type TokenProvider func(ctx context.Context) (string, error)

// CreateTokenProvider creates a token provider function for the given user
func (a *AuthService) CreateTokenProvider(userID uint) TokenProvider {
	return func(ctx context.Context) (string, error) {
		// Find user
		user, err := a.users.FindByID(ctx, userID)
		if err != nil {
			return "", fmt.Errorf("failed to find user: %w", err)
		}

		// Check if user can still get tokens
		if !user.CanLogin() {
			return "", domain.ErrUserNotConfirmed
		}

		// Load user role
		role, err := a.roles.FindByID(ctx, user.RoleID)
		if err != nil {
			return "", fmt.Errorf("failed to load user role: %w", err)
		}
		user.Role = role

		// Generate access token
		now := time.Now()
		accessToken, err := a.tokens.SignAccessToken(jwt.MapClaims{
			"sub":   user.ID,
			"email": user.Email.String(),
			"role":  user.RoleID,
			"type":  "access",
			"exp":   now.Add(a.tokens.GetAccessTokenTTL()).Unix(),
			"iat":   now.Unix(),
		})
		if err != nil {
			return "", fmt.Errorf("failed to sign access token: %w", err)
		}

		return accessToken, nil
	}
}

// LoginWithTokenProvider authenticates a user and returns a token provider
func (a *AuthService) LoginWithTokenProvider(ctx context.Context, req LoginRequest) (*AuthResponse, TokenProvider, error) {
	a.logger.Info("User login with token provider", "email", req.Email)

	// Perform standard login
	authResponse, err := a.Login(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	// Create token provider
	tokenProvider := a.CreateTokenProvider(authResponse.User.ID)

	return authResponse, tokenProvider, nil
}

// ValidateTokenWithStorage validates a token using the token storage if available
func (a *AuthService) ValidateTokenWithStorage(ctx context.Context, tokenStr string) (uint, error) {
	// First validate the JWT token
	claims, err := a.tokens.ValidateAccessToken(tokenStr)
	if err != nil {
		return 0, fmt.Errorf("invalid access token: %w", err)
	}

	// Extract user ID from claims
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return 0, errors.New("invalid token claims")
	}
	userID := uint(userIDFloat)

	// If token storage is available, check if token is revoked
	if a.tokenStorage != nil {
		// Extract token ID from claims (if available)
		if jti, ok := claims["jti"].(string); ok {
			revoked, err := a.tokenStorage.IsTokenRevoked(ctx, jti)
			if err != nil {
				a.logger.Warn("Failed to check token revocation", "error", err)
				// Continue with validation - don't fail on storage errors
			} else if revoked {
				return 0, errors.New("token has been revoked")
			}
		}
	}

	return userID, nil
}

// RevokeToken revokes a specific token using token storage
func (a *AuthService) RevokeToken(ctx context.Context, tokenStr string) error {
	if a.tokenStorage == nil {
		return errors.New("token storage not available")
	}

	// Validate token to get claims
	claims, err := a.tokens.ValidateAccessToken(tokenStr)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Extract token ID from claims
	jti, ok := claims["jti"].(string)
	if !ok {
		return errors.New("token does not have ID")
	}

	// Revoke token in storage
	if err := a.tokenStorage.RevokeToken(ctx, jti); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	a.logger.Info("Token revoked", "tokenId", jti)
	return nil
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (a *AuthService) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	if a.tokenStorage == nil {
		return errors.New("token storage not available")
	}

	// Revoke all tokens in storage
	if err := a.tokenStorage.RevokeAllUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	// Also revoke refresh tokens in database
	if err := a.refreshTokens.DeleteByUserID(ctx, userID); err != nil {
		a.logger.Warn("Failed to revoke refresh tokens", "userId", userID, "error", err)
		// Don't fail the operation
	}

	a.logger.Info("All user tokens revoked", "userId", userID)
	return nil
}

// generateTokenPairWithStorage creates tokens and stores them if storage is available
func (a *AuthService) generateTokenPairWithStorage(ctx context.Context, user *domain.User) (string, string, error) {
	now := time.Now()
	tokenID := fmt.Sprintf("token_%d_%d", user.ID, now.Unix())

	// Generate access token with token ID
	accessToken, err := a.tokens.SignAccessToken(jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email.String(),
		"role":  user.RoleID,
		"type":  "access",
		"exp":   now.Add(a.tokens.GetAccessTokenTTL()).Unix(),
		"iat":   now.Unix(),
		"jti":   tokenID, // Token ID for revocation
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Store token in token storage if available
	if a.tokenStorage != nil {
		if err := a.tokenStorage.StoreToken(ctx, tokenID, user.ID, a.tokens.GetAccessTokenTTL()); err != nil {
			a.logger.Warn("Failed to store token", "tokenId", tokenID, "error", err)
			// Continue - don't fail token generation
		}
	}

	// Create refresh token domain entity
	refreshToken, rawToken, err := domain.NewRefreshToken(user.ID, a.tokens.GetRefreshTokenTTL())
	if err != nil {
		return "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	// Persist refresh token
	if err := a.refreshTokens.Create(ctx, refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to persist refresh token: %w", err)
	}

	// Generate JWT refresh token
	refreshTokenJWT, err := a.tokens.SignRefreshToken(jwt.MapClaims{
		"sub":  user.ID,
		"type": "refresh",
		"exp":  refreshToken.ExpiresAt.Unix(),
		"iat":  now.Unix(),
		"jti":  rawToken, // Use raw token as JWT ID
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshTokenJWT, nil
}

// Delegate to original AuthUseCase methods for backward compatibility
func (a *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	return a.authUseCase.Register(ctx, req)
}

func (a *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	return a.authUseCase.Login(ctx, req)
}

func (a *AuthService) Confirm(ctx context.Context, req ConfirmRequest) (*AuthResponse, error) {
	return a.authUseCase.Confirm(ctx, req)
}

func (a *AuthService) Refresh(ctx context.Context, req RefreshRequest) (*AuthResponse, error) {
	return a.authUseCase.Refresh(ctx, req)
}

func (a *AuthService) Logout(ctx context.Context, req LogoutRequest) error {
	return a.authUseCase.Logout(ctx, req)
}

func (a *AuthService) LogoutAll(ctx context.Context, userID uint) error {
	// Use enhanced logout that also revokes tokens in storage
	return a.RevokeAllUserTokens(ctx, userID)
}

func (a *AuthService) ValidateAccessToken(ctx context.Context, tokenStr string) (uint, error) {
	// Use enhanced validation with storage
	return a.ValidateTokenWithStorage(ctx, tokenStr)
}

func (a *AuthService) GetUserProfile(ctx context.Context, userID uint) (*domain.User, error) {
	return a.authUseCase.GetUserProfile(ctx, userID)
}

func (a *AuthService) ResendConfirmation(ctx context.Context, req ResendConfirmationRequest) error {
	return a.authUseCase.ResendConfirmation(ctx, req)
}
