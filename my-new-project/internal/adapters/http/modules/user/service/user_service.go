// @kthulu:service:user
package service

import (
	"my-new-project/internal/adapters/http/modules/user/domain"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) domain.UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(entity *domain.User) error {
	return s.repo.Create(entity)
}

func (s *UserService) GetUserByID(id uint) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) UpdateUser(entity *domain.User) error {
	return s.repo.Update(entity)
}

func (s *UserService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}

func (s *UserService) ListUsers() ([]*domain.User, error) {
	return s.repo.List()
}