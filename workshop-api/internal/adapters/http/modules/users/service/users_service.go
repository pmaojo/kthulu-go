// @kthulu:service:users
package service

import (
	"github.com/example/workshop-api/internal/adapters/http/modules/users/domain"
)

type UsersService struct {
	repo domain.UsersRepository
}

func NewUsersService(repo domain.UsersRepository) domain.UsersService {
	return &UsersService{repo: repo}
}

func (s *UsersService) CreateUsers(entity *domain.Users) error {
	return s.repo.Create(entity)
}

func (s *UsersService) GetUsersByID(id uint) (*domain.Users, error) {
	return s.repo.GetByID(id)
}

func (s *UsersService) UpdateUsers(entity *domain.Users) error {
	return s.repo.Update(entity)
}

func (s *UsersService) DeleteUsers(id uint) error {
	return s.repo.Delete(id)
}

func (s *UsersService) ListUserses() ([]*domain.Users, error) {
	return s.repo.List()
}