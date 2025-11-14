// @kthulu:module:auth
package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"backend/core"
	"backend/internal/domain"
	"backend/internal/repository"
)

// Mock implementations for testing
type mockUserRepository struct {
	users map[string]*domain.User
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	user.ID = uint(len(m.users) + 1)
	m.users[user.Email.String()] = user
	return nil
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	m.users[user.Email.String()] = user
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id uint) error { return nil }
func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepository) Count(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockUserRepository) FindByRole(ctx context.Context, roleID uint) ([]*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepository) FindUnconfirmed(ctx context.Context, olderThan *time.Time) ([]*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepository) FindPaginated(ctx context.Context, params repository.PaginationParams) (repository.PaginationResult[*domain.User], error) {
	return repository.PaginationResult[*domain.User]{}, nil
}
func (m *mockUserRepository) SearchPaginated(ctx context.Context, query string, params repository.PaginationParams) (repository.PaginationResult[*domain.User], error) {
	return repository.PaginationResult[*domain.User]{}, nil
}
func (m *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, exists := m.users[email]
	return exists, nil
}
func (m *mockUserRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	return false, nil
}

type mockRefreshTokenRepository struct {
	tokens map[string]*domain.RefreshToken
}

func (m *mockRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	token.ID = uint(len(m.tokens) + 1)
	if m.tokens == nil {
		m.tokens = make(map[string]*domain.RefreshToken)
	}
	m.tokens[token.Token] = token
	return nil
}

func (m *mockRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	hashed := hashToken(token)
	if t, exists := m.tokens[hashed]; exists {
		return t, nil
	}
	return nil, domain.ErrTokenNotFound
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (m *mockRefreshTokenRepository) Delete(ctx context.Context, id uint) error {
	for token, t := range m.tokens {
		if t.ID == id {
			delete(m.tokens, token)
			break
		}
	}
	return nil
}

func (m *mockRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	for token, t := range m.tokens {
		if t.UserID == userID {
			delete(m.tokens, token)
		}
	}
	return nil
}

// Implement other required methods with no-ops
func (m *mockRefreshTokenRepository) FindByID(ctx context.Context, id uint) (*domain.RefreshToken, error) {
	return nil, nil
}
func (m *mockRefreshTokenRepository) Update(ctx context.Context, token *domain.RefreshToken) error {
	return nil
}
func (m *mockRefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	return nil
}
func (m *mockRefreshTokenRepository) FindByUserID(ctx context.Context, userID uint) ([]*domain.RefreshToken, error) {
	return nil, nil
}
func (m *mockRefreshTokenRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	return 0, nil
}
func (m *mockRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockRefreshTokenRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	return 0, nil
}
func (m *mockRefreshTokenRepository) List(ctx context.Context, limit, offset int) ([]*domain.RefreshToken, error) {
	return nil, nil
}
func (m *mockRefreshTokenRepository) Count(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockRefreshTokenRepository) FindExpired(ctx context.Context) ([]*domain.RefreshToken, error) {
	return nil, nil
}
func (m *mockRefreshTokenRepository) ExistsByToken(ctx context.Context, token string) (bool, error) {
	return false, nil
}
func (m *mockRefreshTokenRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	return false, nil
}
func (m *mockRefreshTokenRepository) IsValidToken(ctx context.Context, token string) (bool, error) {
	return false, nil
}

type mockRoleRepository struct {
	roles map[string]*domain.Role
}

func (m *mockRoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	if role, exists := m.roles[name]; exists {
		return role, nil
	}
	return nil, domain.ErrRoleNotFound
}

func (m *mockRoleRepository) FindByID(ctx context.Context, id uint) (*domain.Role, error) {
	for _, role := range m.roles {
		if role.ID == id {
			return role, nil
		}
	}
	return nil, domain.ErrRoleNotFound
}

func (m *mockRoleRepository) FindByUserID(ctx context.Context, userID uint) (*domain.Role, error) {
	return m.FindByID(ctx, userID)
}

func (m *mockRoleRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	for _, role := range m.roles {
		if role.ID == id {
			return true, nil
		}
	}
	return false, nil
}

// Implement other required methods with no-ops
func (m *mockRoleRepository) Create(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepository) Update(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockRoleRepository) Delete(ctx context.Context, id uint) error           { return nil }
func (m *mockRoleRepository) List(ctx context.Context) ([]*domain.Role, error)    { return nil, nil }
func (m *mockRoleRepository) Count(ctx context.Context) (int64, error)            { return 0, nil }
func (m *mockRoleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	return false, nil
}
func (m *mockRoleRepository) AddPermission(ctx context.Context, roleID, permissionID uint) error {
	return nil
}
func (m *mockRoleRepository) RemovePermission(ctx context.Context, roleID, permissionID uint) error {
	return nil
}
func (m *mockRoleRepository) GetRolePermissions(ctx context.Context, roleID uint) ([]*domain.Permission, error) {
	return nil, nil
}

type mockTokenManager struct{}

func (m *mockTokenManager) SignAccessToken(claims jwt.Claims) (string, error) {
	return "mock-access-token", nil
}

func (m *mockTokenManager) SignRefreshToken(claims jwt.Claims) (string, error) {
	return "mock-refresh-token", nil
}

func (m *mockTokenManager) ValidateAccessToken(token string) (jwt.MapClaims, error) {
	return jwt.MapClaims{"sub": float64(1)}, nil
}

func (m *mockTokenManager) ValidateRefreshToken(token string) (jwt.MapClaims, error) {
	return jwt.MapClaims{"sub": float64(1)}, nil
}

func (m *mockTokenManager) GetAccessTokenTTL() time.Duration {
	return 15 * time.Minute
}

func (m *mockTokenManager) GetRefreshTokenTTL() time.Duration {
	return 7 * 24 * time.Hour
}

type mockNotificationProvider struct{}

func (m *mockNotificationProvider) SendNotification(ctx context.Context, req repository.NotificationRequest) error {
	return nil
}

func (m *mockNotificationProvider) SendEmailConfirmation(ctx context.Context, email, confirmationCode string) error {
	return nil
}

func (m *mockNotificationProvider) SendPasswordReset(ctx context.Context, email, resetCode string) error {
	return nil
}

func (m *mockNotificationProvider) SendWelcomeEmail(ctx context.Context, email, name string) error {
	return nil
}

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...interface{}) {}
func (m *mockLogger) Info(msg string, fields ...interface{})  {}
func (m *mockLogger) Warn(msg string, fields ...interface{})  {}
func (m *mockLogger) Error(msg string, fields ...interface{}) {}
func (m *mockLogger) Fatal(msg string, fields ...interface{}) {}
func (m *mockLogger) With(fields ...interface{}) core.Logger  { return m }
func (m *mockLogger) Sync() error                             { return nil }

func TestAuthUseCase_Register(t *testing.T) {
	// Setup mocks
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	refreshTokenRepo := &mockRefreshTokenRepository{tokens: make(map[string]*domain.RefreshToken)}
	roleRepo := &mockRoleRepository{roles: make(map[string]*domain.Role)}
	tokenManager := &mockTokenManager{}
	notifier := &mockNotificationProvider{}
	logger := &mockLogger{}

	// Create default role
	defaultRole, _ := domain.NewRole(domain.RoleUser, "Default user role")
	defaultRole.ID = 1
	roleRepo.roles[domain.RoleUser] = defaultRole

	// Create auth use case
	authUC := NewAuthUseCase(userRepo, refreshTokenRepo, roleRepo, tokenManager, notifier, logger)

	// Test registration
	req := RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	response, err := authUC.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	if response.User == nil {
		t.Fatal("Expected user in response")
	}

	if response.User.Email.String() != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, response.User.Email.String())
	}

	// User should not be confirmed initially
	if response.User.IsConfirmed() {
		t.Error("User should not be confirmed initially")
	}

	// Should not have tokens until confirmed
	if response.AccessToken != "" || response.RefreshToken != "" {
		t.Error("Should not have tokens until email is confirmed")
	}
}

func TestAuthUseCase_Login_UnconfirmedUser(t *testing.T) {
	// Setup mocks
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	refreshTokenRepo := &mockRefreshTokenRepository{tokens: make(map[string]*domain.RefreshToken)}
	roleRepo := &mockRoleRepository{roles: make(map[string]*domain.Role)}
	tokenManager := &mockTokenManager{}
	notifier := &mockNotificationProvider{}
	logger := &mockLogger{}

	// Create default role
	defaultRole, _ := domain.NewRole(domain.RoleUser, "Default user role")
	defaultRole.ID = 1
	roleRepo.roles[domain.RoleUser] = defaultRole

	// Create auth use case
	authUC := NewAuthUseCase(userRepo, refreshTokenRepo, roleRepo, tokenManager, notifier, logger)

	// Register user first
	registerReq := RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := authUC.Register(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Try to login with unconfirmed user
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	_, err = authUC.Login(context.Background(), loginReq)
	if err != domain.ErrUserNotConfirmed {
		t.Errorf("Expected ErrUserNotConfirmed, got %v", err)
	}
}

func TestAuthUseCase_Confirm(t *testing.T) {
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	refreshTokenRepo := &mockRefreshTokenRepository{tokens: make(map[string]*domain.RefreshToken)}
	roleRepo := &mockRoleRepository{roles: make(map[string]*domain.Role)}
	tokenManager := &mockTokenManager{}
	notifier := &mockNotificationProvider{}
	logger := &mockLogger{}

	defaultRole, _ := domain.NewRole(domain.RoleUser, "Default user role")
	defaultRole.ID = 1
	roleRepo.roles[domain.RoleUser] = defaultRole

	authUC := NewAuthUseCase(userRepo, refreshTokenRepo, roleRepo, tokenManager, notifier, logger)

	registerReq := RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	if _, err := authUC.Register(context.Background(), registerReq); err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	storedUser := userRepo.users[registerReq.Email]
	if storedUser == nil || storedUser.ConfirmationCode == "" {
		t.Fatalf("Expected confirmation code to be stored")
	}

	// Successful confirmation
	confirmReq := ConfirmRequest{Email: registerReq.Email, ConfirmationCode: storedUser.ConfirmationCode}
	resp, err := authUC.Confirm(context.Background(), confirmReq)
	if err != nil {
		t.Fatalf("Confirmation failed: %v", err)
	}
	if !resp.User.IsConfirmed() {
		t.Fatalf("User should be confirmed")
	}
	if userRepo.users[registerReq.Email].ConfirmationCode != "" {
		t.Fatalf("Confirmation code should be invalidated after confirmation")
	}

	// Invalid confirmation code
	registerReq2 := RegisterRequest{Email: "test2@example.com", Password: "password123"}
	if _, err := authUC.Register(context.Background(), registerReq2); err != nil {
		t.Fatalf("Registration failed: %v", err)
	}
	badReq := ConfirmRequest{Email: registerReq2.Email, ConfirmationCode: "wrong"}
	if _, err := authUC.Confirm(context.Background(), badReq); err == nil {
		t.Fatalf("Expected error for invalid confirmation code")
	}
	if userRepo.users[registerReq2.Email].IsConfirmed() {
		t.Fatalf("User should not be confirmed with invalid code")
	}
}
