package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"{{.ModuleName}}/internal/modules/auth/domain"
)

type AuthService struct {
	repo      domain.AuthRepository
	jwtSecret string
}

func NewAuthService(repo domain.AuthRepository) domain.AuthService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "CHANGE_ME_IN_PRODUCTION" // Fallback for dev
	}
	return &AuthService{
		repo:      repo,
		jwtSecret: secret,
	}
}

func (s *AuthService) Register(email, password string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:    email,
		Password: string(hash),
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}
