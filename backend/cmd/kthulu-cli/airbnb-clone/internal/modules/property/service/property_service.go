// @kthulu:service:property
package service

import (
	"airbnb-clone/internal/modules/property/domain"
)

type PropertyService struct {
	repo domain.PropertyRepository
}

func NewPropertyService(repo domain.PropertyRepository) domain.PropertyService {
	return &PropertyService{repo: repo}
}

func (s *PropertyService) CreateProperty(entity *domain.Property) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *PropertyService) GetPropertyByID(id uint) (*domain.Property, error) {
	return s.repo.GetByID(id)
}

func (s *PropertyService) UpdateProperty(entity *domain.Property) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *PropertyService) DeleteProperty(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *PropertyService) ListProperties(filter domain.SearchFilter) ([]*domain.Property, error) {
	return s.repo.List(filter)
}
