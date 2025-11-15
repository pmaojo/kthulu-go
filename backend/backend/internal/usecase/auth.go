// @kthulu:module:auth
package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// AuthUseCase orchestrates user authentication workflows.
type AuthUseCase struct {
	users         repository.UserRepository
	refreshTokens repository.RefreshTokenRepository
	roles         repository.RoleRepository
	tokens        core.TokenManager
	notifier      repository.NotificationProvider
	logger        core.Logger
}

// NewAuthUseCase builds an AuthUseCase instance.
func NewAuthUseCase(
	users repository.UserRepository,
	refreshTokens repository.RefreshTokenRepository,
	roles repository.RoleRepository,
	tokens core.TokenManager,
	notifier repository.NotificationProvider,
	logger core.Logger,
) *AuthUseCase {
	return &AuthUseCase{
		users:         users,
		refreshTokens: refreshTokens,
		roles:         roles,
		tokens:        tokens,
		notifier:      notifier,
		logger:        logger,
	}
}

// RegisterRequest contains the data needed to register a new user
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	RoleID   uint   `json:"roleId,omitempty"`
}

// AuthResponse contains the authentication response data
type AuthResponse struct {
	User         *domain.User `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	ExpiresIn    int64        `json:"expiresIn"`
}

// Register creates a new user and returns access and refresh tokens.
// The user will need to confirm their email before they can login.
func (a *AuthUseCase) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	ctx, span := startUseCaseSpan(ctx, "AuthUseCase.Register")
	defer span.End()

	a.logger.Info("Starting user registration", "email", req.Email)

	// Validate password strength
	if err := a.validatePassword(req.Password); err != nil {
		a.logger.Warn("Password validation failed", "email", req.Email, "error", err)
		return nil, err
	}

	// Check if user already exists
	exists, err := a.users.ExistsByEmail(ctx, req.Email)
	if err != nil {
		a.logger.Error("Failed to check user existence", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		a.logger.Warn("User registration attempted for existing email", "email", req.Email)
		return nil, domain.ErrUserAlreadyExists
	}

	// Set default role if not provided
	roleID := req.RoleID
	if roleID == 0 {
		defaultRole, err := a.roles.FindByName(ctx, domain.RoleUser)
		if err != nil {
			a.logger.Error("Failed to find default role", "error", err)
			return nil, fmt.Errorf("failed to find default role: %w", err)
		}
		roleID = defaultRole.ID
	}

	// Validate role exists
	roleExists, err := a.roles.ExistsByID(ctx, roleID)
	if err != nil {
		a.logger.Error("Failed to validate role", "roleId", roleID, "error", err)
		return nil, fmt.Errorf("failed to validate role: %w", err)
	}
	if !roleExists {
		a.logger.Warn("Registration attempted with invalid role", "roleId", roleID)
		return nil, domain.ErrInvalidRole
	}

	// Hash password
	hashed, err := a.hashPassword(req.Password)
	if err != nil {
		a.logger.Error("Failed to hash password", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user domain entity (unconfirmed by default)
	user, err := domain.NewUser(req.Email, hashed, roleID)
	if err != nil {
		a.logger.Error("Failed to create user domain entity", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate confirmation code and set on user
	confirmationCode, err := a.GenerateConfirmationCode()
	if err != nil {
		a.logger.Error("Failed to generate confirmation code", "email", req.Email, "error", err)
		// Don't fail registration if code generation fails
	} else {
		user.ConfirmationCode = confirmationCode
	}

	// Persist user
	if err := a.users.Create(ctx, user); err != nil {
		a.logger.Error("Failed to persist user", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Send confirmation email if we have a code
	if confirmationCode != "" {
		if err := a.notifier.SendEmailConfirmation(ctx, req.Email, confirmationCode); err != nil {
			a.logger.Error("Failed to send confirmation email", "userId", user.ID, "email", req.Email, "error", err)
			// Don't fail registration if email sending fails
		} else {
			a.logger.Info("Confirmation email sent", "userId", user.ID, "email", req.Email)
		}
	}

	a.logger.Info("User registered successfully", "userId", user.ID, "email", req.Email)

	// Note: For registration, we don't immediately provide tokens
	// The user needs to confirm their email first
	return &AuthResponse{
		User: user,
		// No tokens until email is confirmed
	}, nil
}

// LoginRequest contains the data needed to authenticate a user
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Login authenticates an existing user and returns access and refresh tokens.
func (a *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	ctx, span := startUseCaseSpan(ctx, "AuthUseCase.Login")
	defer span.End()

	a.logger.Info("User login attempt", "email", req.Email)

	// Find user by email
	user, err := a.users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			a.logger.Warn("Login attempt for non-existent user", "email", req.Email)
			return nil, domain.ErrUserNotFound
		}
		a.logger.Error("Failed to find user during login", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user can login (must be confirmed)
	if !user.CanLogin() {
		a.logger.Warn("Login attempt for unconfirmed user", "userId", user.ID, "email", req.Email)
		return nil, domain.ErrUserNotConfirmed
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		a.logger.Warn("Invalid password during login", "email", req.Email)
		return nil, domain.ErrUserNotFound // Don't reveal if user exists
	}

	// Load user role for token claims
	role, err := a.roles.FindByID(ctx, user.RoleID)
	if err != nil {
		a.logger.Error("Failed to load user role", "userId", user.ID, "roleId", user.RoleID, "error", err)
		return nil, fmt.Errorf("failed to load user role: %w", err)
	}
	user.Role = role

	// Generate tokens
	accessToken, refreshTokenStr, err := a.generateTokenPair(ctx, user)
	if err != nil {
		a.logger.Error("Failed to generate tokens", "userId", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	a.logger.Info("User logged in successfully", "userId", user.ID, "email", req.Email)

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    int64(a.tokens.GetAccessTokenTTL().Seconds()),
	}, nil
}

// ConfirmRequest contains the data needed to confirm a user's email
type ConfirmRequest struct {
	Email            string `json:"email" validate:"required,email"`
	ConfirmationCode string `json:"confirmationCode" validate:"required"`
}

// Confirm marks a user as confirmed and returns authentication tokens.
func (a *AuthUseCase) Confirm(ctx context.Context, req ConfirmRequest) (*AuthResponse, error) {
	ctx, span := startUseCaseSpan(ctx, "AuthUseCase.Confirm")
	defer span.End()

	a.logger.Info("User email confirmation attempt", "email", req.Email)

	// Find user by email
	user, err := a.users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			a.logger.Warn("Confirmation attempt for non-existent user", "email", req.Email)
			return nil, domain.ErrUserNotFound
		}
		a.logger.Error("Failed to find user during confirmation", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is already confirmed
	if user.IsConfirmed() {
		a.logger.Info("User already confirmed", "userId", user.ID, "email", req.Email)
		// Still return success but don't generate new tokens
		return &AuthResponse{User: user}, nil
	}

	// Validate the confirmation code
	if req.ConfirmationCode == "" || req.ConfirmationCode != user.ConfirmationCode {
		a.logger.Warn("Invalid confirmation code provided", "email", req.Email)
		return nil, errors.New("invalid confirmation code")
	}

	// Use domain method to confirm user (also clears confirmation code)
	user.Confirm()

	// Update user in database
	if err := a.users.Update(ctx, user); err != nil {
		a.logger.Error("Failed to update user confirmation", "userId", user.ID, "error", err)
		return nil, fmt.Errorf("failed to confirm user: %w", err)
	}

	// Load user role
	role, err := a.roles.FindByID(ctx, user.RoleID)
	if err != nil {
		a.logger.Error("Failed to load user role after confirmation", "userId", user.ID, "roleId", user.RoleID, "error", err)
		return nil, fmt.Errorf("failed to load user role: %w", err)
	}
	user.Role = role

	// Generate tokens for the newly confirmed user
	accessToken, refreshTokenStr, err := a.generateTokenPair(ctx, user)
	if err != nil {
		a.logger.Error("Failed to generate tokens after confirmation", "userId", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	a.logger.Info("User email confirmed successfully", "userId", user.ID, "email", req.Email)

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    int64(a.tokens.GetAccessTokenTTL().Seconds()),
	}, nil
}

// RefreshRequest contains the data needed to refresh tokens
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// Refresh validates a refresh token and returns new access and refresh tokens.
func (a *AuthUseCase) Refresh(ctx context.Context, req RefreshRequest) (*AuthResponse, error) {
	ctx, span := startUseCaseSpan(ctx, "AuthUseCase.Refresh")
	defer span.End()

	a.logger.Info("Token refresh attempt")

	// Validate the JWT refresh token
	claims, err := a.tokens.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		a.logger.Warn("Invalid refresh token provided", "error", err)
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Extract user ID from claims
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		a.logger.Warn("Invalid user ID in refresh token claims")
		return nil, errors.New("invalid token claims")
	}
	userID := uint(userIDFloat)

	jti, ok := claims["jti"].(string)
	if !ok {
		a.logger.Warn("Refresh token missing jti")
		return nil, domain.ErrInvalidToken
	}

	// Check if refresh token exists in database
	refreshToken, err := a.refreshTokens.FindByToken(ctx, jti)
	if err != nil {
		if errors.Is(err, domain.ErrTokenNotFound) {
			a.logger.Warn("Refresh token not found in database", "userId", userID)
			return nil, domain.ErrInvalidToken
		}
		a.logger.Error("Failed to find refresh token", "userId", userID, "error", err)
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Check if token is expired
	if refreshToken.IsExpired() {
		a.logger.Warn("Expired refresh token used", "userId", userID, "tokenId", refreshToken.ID)
		// Clean up expired token
		_ = a.refreshTokens.Delete(ctx, refreshToken.ID)
		return nil, domain.ErrTokenExpired
	}

	// Find user
	user, err := a.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			a.logger.Warn("Refresh token for non-existent user", "userId", userID)
			return nil, domain.ErrUserNotFound
		}
		a.logger.Error("Failed to find user during token refresh", "userId", userID, "error", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user can still login
	if !user.CanLogin() {
		a.logger.Warn("Refresh token used for unconfirmed user", "userId", userID)
		return nil, domain.ErrUserNotConfirmed
	}

	// Load user role
	role, err := a.roles.FindByID(ctx, user.RoleID)
	if err != nil {
		a.logger.Error("Failed to load user role during refresh", "userId", userID, "roleId", user.RoleID, "error", err)
		return nil, fmt.Errorf("failed to load user role: %w", err)
	}
	user.Role = role

	// Delete old refresh token
	if err := a.refreshTokens.Delete(ctx, refreshToken.ID); err != nil {
		a.logger.Error("Failed to delete old refresh token", "tokenId", refreshToken.ID, "error", err)
		// Continue anyway, as this is not critical
	}

	// Generate new token pair
	accessToken, newRefreshTokenStr, err := a.generateTokenPair(ctx, user)
	if err != nil {
		a.logger.Error("Failed to generate new tokens", "userId", userID, "error", err)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	a.logger.Info("Tokens refreshed successfully", "userId", userID)

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenStr,
		ExpiresIn:    int64(a.tokens.GetAccessTokenTTL().Seconds()),
	}, nil
}

// LogoutRequest contains the data needed to logout a user
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// Logout invalidates the user's refresh token.
func (a *AuthUseCase) Logout(ctx context.Context, req LogoutRequest) error {
	ctx, span := startUseCaseSpan(ctx, "AuthUseCase.Logout")
	defer span.End()

	a.logger.Info("User logout attempt")

	// Validate token and extract jti
	claims, err := a.tokens.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		a.logger.Error("Failed to validate refresh token during logout", "error", err)
		return fmt.Errorf("failed to validate refresh token: %w", err)
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		a.logger.Warn("Refresh token missing jti during logout")
		return domain.ErrInvalidToken
	}

	// Find and delete the refresh token
	refreshToken, err := a.refreshTokens.FindByToken(ctx, jti)
	if err != nil {
		if errors.Is(err, domain.ErrTokenNotFound) {
			// Token doesn't exist, consider logout successful
			a.logger.Info("Logout with non-existent token")
			return nil
		}
		a.logger.Error("Failed to find refresh token during logout", "error", err)
		return fmt.Errorf("failed to find refresh token: %w", err)
	}

	// Delete the refresh token
	if err := a.refreshTokens.Delete(ctx, refreshToken.ID); err != nil {
		a.logger.Error("Failed to delete refresh token during logout", "tokenId", refreshToken.ID, "error", err)
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	a.logger.Info("User logged out successfully", "userId", refreshToken.UserID)
	return nil
}

// LogoutAll invalidates all refresh tokens for a user.
func (a *AuthUseCase) LogoutAll(ctx context.Context, userID uint) error {
	a.logger.Info("User logout all attempt", "userId", userID)

	// Delete all refresh tokens for the user
	if err := a.refreshTokens.DeleteByUserID(ctx, userID); err != nil {
		a.logger.Error("Failed to delete all refresh tokens", "userId", userID, "error", err)
		return fmt.Errorf("failed to delete refresh tokens: %w", err)
	}

	a.logger.Info("All user sessions logged out", "userId", userID)
	return nil
}

// generateTokenPair creates both access and refresh tokens for a user
func (a *AuthUseCase) generateTokenPair(ctx context.Context, user *domain.User) (string, string, error) {
	now := time.Now()

	// Generate access token
	accessToken, err := a.tokens.SignAccessToken(jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email.String(),
		"role":  user.RoleID,
		"type":  "access",
		"exp":   now.Add(a.tokens.GetAccessTokenTTL()).Unix(),
		"iat":   now.Unix(),
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token domain entity
	refreshToken, rawToken, err := domain.NewRefreshToken(user.ID, a.tokens.GetRefreshTokenTTL())
	if err != nil {
		return "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	// Persist refresh token (only hash stored)
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

// hashPassword hashes a password using bcrypt
func (a *AuthUseCase) hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashed), nil
}

// validatePassword validates password strength
func (a *AuthUseCase) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Add more password validation rules as needed
	// For example: require uppercase, lowercase, numbers, special characters

	return nil
}

// ResendConfirmationRequest contains the data needed to resend confirmation email
type ResendConfirmationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResendConfirmation sends a new confirmation email to an unconfirmed user
func (a *AuthUseCase) ResendConfirmation(ctx context.Context, req ResendConfirmationRequest) error {
	ctx, span := startUseCaseSpan(ctx, "AuthUseCase.ResendConfirmation")
	defer span.End()

	a.logger.Info("Resend confirmation email request", "email", req.Email)

	// Find user by email
	user, err := a.users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			a.logger.Warn("Resend confirmation for non-existent user", "email", req.Email)
			// Don't reveal if user exists
			return nil
		}
		a.logger.Error("Failed to find user for confirmation resend", "email", req.Email, "error", err)
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is already confirmed
	if user.IsConfirmed() {
		a.logger.Info("Resend confirmation for already confirmed user", "userId", user.ID, "email", req.Email)
		// Don't reveal if user is already confirmed
		return nil
	}

	// Generate new confirmation code
	confirmationCode, err := a.GenerateConfirmationCode()
	if err != nil {
		a.logger.Error("Failed to generate confirmation code for resend", "userId", user.ID, "error", err)
		return fmt.Errorf("failed to generate confirmation code: %w", err)
	}

	user.ConfirmationCode = confirmationCode
	if err := a.users.Update(ctx, user); err != nil {
		a.logger.Error("Failed to store new confirmation code", "userId", user.ID, "error", err)
		return fmt.Errorf("failed to store confirmation code: %w", err)
	}

	// Send confirmation email
	if err := a.notifier.SendEmailConfirmation(ctx, req.Email, confirmationCode); err != nil {
		a.logger.Error("Failed to resend confirmation email", "userId", user.ID, "email", req.Email, "error", err)
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	a.logger.Info("Confirmation email resent", "userId", user.ID, "email", req.Email)
	return nil
}

// GetUserProfile retrieves user profile information
func (a *AuthUseCase) GetUserProfile(ctx context.Context, userID uint) (*domain.User, error) {
	a.logger.Info("Get user profile request", "userId", userID)

	// Find user by ID
	user, err := a.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			a.logger.Warn("Profile request for non-existent user", "userId", userID)
			return nil, domain.ErrUserNotFound
		}
		a.logger.Error("Failed to find user for profile", "userId", userID, "error", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Load user role
	role, err := a.roles.FindByID(ctx, user.RoleID)
	if err != nil {
		a.logger.Error("Failed to load user role for profile", "userId", userID, "roleId", user.RoleID, "error", err)
		return nil, fmt.Errorf("failed to load user role: %w", err)
	}
	user.Role = role

	a.logger.Info("User profile retrieved", "userId", userID)
	return user, nil
}

// ValidateAccessToken validates an access token and returns the user ID
func (a *AuthUseCase) ValidateAccessToken(ctx context.Context, tokenStr string) (uint, error) {
	claims, err := a.tokens.ValidateAccessToken(tokenStr)
	if err != nil {
		return 0, fmt.Errorf("invalid access token: %w", err)
	}

	// Extract user ID from claims
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	return uint(userIDFloat), nil
}

// GenerateConfirmationCode generates a secure confirmation code for email verification
func (a *AuthUseCase) GenerateConfirmationCode() (string, error) {
	bytes := make([]byte, 16) // 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate confirmation code: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
