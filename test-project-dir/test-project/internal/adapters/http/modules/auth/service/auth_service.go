// @kthulu:service:auth
package service

import (
	"test-project/internal/adapters/http/modules/auth/domain"
)

type AuthService struct {
	repo domain.AuthRepository
}

func NewAuthService(repo domain.AuthRepository) domain.AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) CreateAuth(entity *domain.Auth) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *AuthService) GetAuthByID(id uint) (*domain.Auth, error) {
	return s.repo.GetByID(id)
}

func (s *AuthService) UpdateAuth(entity *domain.Auth) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *AuthService) DeleteAuth(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *AuthsService) ListAuth() ([]*domain.%!s(MISSING), error) {
	return s.repo.List()
}
