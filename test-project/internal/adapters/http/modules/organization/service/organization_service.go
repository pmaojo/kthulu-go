// @kthulu:service:organization
package service

import (
	"test-project/internal/adapters/http/modules/organization/domain"
)

type OrganizationService struct {
	repo domain.OrganizationRepository
}

func NewOrganizationService(repo domain.OrganizationRepository) domain.OrganizationService {
	return &OrganizationService{repo: repo}
}

func (s *OrganizationService) CreateOrganization(entity *domain.Organization) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *OrganizationService) GetOrganizationByID(id uint) (*domain.Organization, error) {
	return s.repo.GetByID(id)
}

func (s *OrganizationService) UpdateOrganization(entity *domain.Organization) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *OrganizationService) DeleteOrganization(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *OrganizationsService) ListOrganization() ([]*domain.%!s(MISSING), error) {
	return s.repo.List()
}
