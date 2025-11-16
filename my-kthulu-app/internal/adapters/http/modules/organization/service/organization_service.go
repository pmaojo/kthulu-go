// @kthulu:service:organization
package service

import (
	"my-kthulu-app/internal/adapters/http/modules/organization/domain"
)

type OrganizationService struct {
	repo domain.OrganizationRepository
}

func NewOrganizationService(repo domain.OrganizationRepository) domain.OrganizationService {
	return &OrganizationService{repo: repo}
}

func (s *OrganizationService) CreateOrganization(entity *domain.Organization) error {
	return s.repo.Create(entity)
}

func (s *OrganizationService) GetOrganizationByID(id uint) (*domain.Organization, error) {
	return s.repo.GetByID(id)
}

func (s *OrganizationService) UpdateOrganization(entity *domain.Organization) error {
	return s.repo.Update(entity)
}

func (s *OrganizationService) DeleteOrganization(id uint) error {
	return s.repo.Delete(id)
}

func (s *OrganizationService) ListOrganizations() ([]*domain.Organization, error) {
	return s.repo.List()
}