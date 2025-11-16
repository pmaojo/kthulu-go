package users

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthService provides basic registration and authentication operations.
type AuthService struct {
	repo UserRepository
}

// NewAuthService creates a new AuthService.
func NewAuthService(repo UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Register creates a new user with a hashed password.
func (s *AuthService) Register(ctx context.Context, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	user := &User{Email: email, PasswordHash: string(hash), CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// ErrInvalidCredentials indicates provided credentials are invalid.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Login verifies user credentials and returns the user if valid.
func (s *AuthService) Login(ctx context.Context, email, password string) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// GetByID retrieves a user by its identifier.
func (s *AuthService) GetByID(ctx context.Context, id uint) (*User, error) {
	return s.repo.FindByID(ctx, id)
}
