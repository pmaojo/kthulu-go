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
	return s.repo.Create(entity)
}

func (s *AuthService) GetAuthByID(id uint) (*domain.Auth, error) {
	return s.repo.GetByID(id)
}

func (s *AuthService) UpdateAuth(entity *domain.Auth) error {
	return s.repo.Update(entity)
}

func (s *AuthService) DeleteAuth(id uint) error {
	return s.repo.Delete(id)
}

func (s *AuthService) ListAuths() ([]*domain.Auth, error) {
	return s.repo.List()
}