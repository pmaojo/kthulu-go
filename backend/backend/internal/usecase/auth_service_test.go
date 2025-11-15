package usecase

import (
	"context"
	"testing"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
)

func TestNewAuthService_InitializesAuthUseCase(t *testing.T) {
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	refreshTokenRepo := &mockRefreshTokenRepository{tokens: make(map[string]*domain.RefreshToken)}
	roleRepo := &mockRoleRepository{roles: make(map[string]*domain.Role)}
	tokenManager := &mockTokenManager{}
	notifier := &mockNotificationProvider{}
	logger := &mockLogger{}

	svc := NewAuthService(userRepo, refreshTokenRepo, roleRepo, nil, tokenManager, notifier, logger)

	if svc.authUseCase == nil {
		t.Fatalf("authUseCase should be initialized")
	}
	if svc.authUseCase.users != userRepo {
		t.Fatalf("authUseCase user repository mismatch")
	}
}

func TestAuthService_Register(t *testing.T) {
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	refreshTokenRepo := &mockRefreshTokenRepository{tokens: make(map[string]*domain.RefreshToken)}
	roleRepo := &mockRoleRepository{roles: make(map[string]*domain.Role)}
	tokenManager := &mockTokenManager{}
	notifier := &mockNotificationProvider{}
	logger := &mockLogger{}

	defaultRole, _ := domain.NewRole(domain.RoleUser, "Default user role")
	defaultRole.ID = 1
	roleRepo.roles[domain.RoleUser] = defaultRole

	svc := NewAuthService(userRepo, refreshTokenRepo, roleRepo, nil, tokenManager, notifier, logger)

	req := RegisterRequest{Email: "test@example.com", Password: "password123"}
	resp, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	if resp == nil || resp.User == nil {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAuthService_Login_UnconfirmedUser(t *testing.T) {
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	refreshTokenRepo := &mockRefreshTokenRepository{tokens: make(map[string]*domain.RefreshToken)}
	roleRepo := &mockRoleRepository{roles: make(map[string]*domain.Role)}
	tokenManager := &mockTokenManager{}
	notifier := &mockNotificationProvider{}
	logger := &mockLogger{}

	defaultRole, _ := domain.NewRole(domain.RoleUser, "Default user role")
	defaultRole.ID = 1
	roleRepo.roles[domain.RoleUser] = defaultRole

	svc := NewAuthService(userRepo, refreshTokenRepo, roleRepo, nil, tokenManager, notifier, logger)

	registerReq := RegisterRequest{Email: "test@example.com", Password: "password123"}
	if _, err := svc.Register(context.Background(), registerReq); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	loginReq := LoginRequest{Email: registerReq.Email, Password: registerReq.Password}
	if _, err := svc.Login(context.Background(), loginReq); err != domain.ErrUserNotConfirmed {
		t.Fatalf("expected ErrUserNotConfirmed, got %v", err)
	}
}

func TestAuthService_Confirm(t *testing.T) {
	userRepo := &mockUserRepository{users: make(map[string]*domain.User)}
	refreshTokenRepo := &mockRefreshTokenRepository{tokens: make(map[string]*domain.RefreshToken)}
	roleRepo := &mockRoleRepository{roles: make(map[string]*domain.Role)}
	tokenManager := &mockTokenManager{}
	notifier := &mockNotificationProvider{}
	logger := &mockLogger{}

	defaultRole, _ := domain.NewRole(domain.RoleUser, "Default user role")
	defaultRole.ID = 1
	roleRepo.roles[domain.RoleUser] = defaultRole

	svc := NewAuthService(userRepo, refreshTokenRepo, roleRepo, nil, tokenManager, notifier, logger)

	registerReq := RegisterRequest{Email: "test@example.com", Password: "password123"}
	if _, err := svc.Register(context.Background(), registerReq); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	storedUser := userRepo.users[registerReq.Email]
	if storedUser == nil || storedUser.ConfirmationCode == "" {
		t.Fatalf("expected confirmation code to be stored")
	}

	confirmReq := ConfirmRequest{Email: registerReq.Email, ConfirmationCode: storedUser.ConfirmationCode}
	resp, err := svc.Confirm(context.Background(), confirmReq)
	if err != nil {
		t.Fatalf("confirmation failed: %v", err)
	}
	if !resp.User.IsConfirmed() {
		t.Fatalf("user should be confirmed")
	}
	if userRepo.users[registerReq.Email].ConfirmationCode != "" {
		t.Fatalf("confirmation code should be cleared")
	}

	registerReq2 := RegisterRequest{Email: "test2@example.com", Password: "password123"}
	if _, err := svc.Register(context.Background(), registerReq2); err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	badReq := ConfirmRequest{Email: registerReq2.Email, ConfirmationCode: "wrong"}
	if _, err := svc.Confirm(context.Background(), badReq); err == nil {
		t.Fatalf("expected error for invalid confirmation code")
	}
	if userRepo.users[registerReq2.Email].IsConfirmed() {
		t.Fatalf("user should not be confirmed with invalid code")
	}
}
