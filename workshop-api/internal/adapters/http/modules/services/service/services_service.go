// @kthulu:service:services
package service

import (
        "github.com/example/workshop-api/internal/adapters/http/modules/services/domain"
)

type ServicesService struct {
        repo domain.ServicesRepository
}

func NewServicesService(repo domain.ServicesRepository) domain.ServicesService {
        return &ServicesService{repo: repo}
}

func (s *ServicesService) CreateServices(entity *domain.Services) error {
        // Add business logic here
        return s.repo.Create(entity)
}

func (s *ServicesService) GetServicesByID(id uint) (*domain.Services, error) {
        return s.repo.GetByID(id)
}

func (s *ServicesService) UpdateServices(entity *domain.Services) error {
        // Add business logic here
        return s.repo.Update(entity)
}

func (s *ServicesService) DeleteServices(id uint) error {
        // Add business logic here
        return s.repo.Delete(id)
}

func (s *ServicesService) ListServicess(filter domain.SearchFilter) ([]*domain.Services, error) {
        return s.repo.List(filter)
}
