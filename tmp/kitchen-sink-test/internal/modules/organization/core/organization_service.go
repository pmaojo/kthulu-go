// @kthulu:service:organization
package core

type organizationService struct {
	repo OrganizationRepository
}

func NewOrganizationService(repo OrganizationRepository) OrganizationService {
	return &organizationService{repo: repo}
}

func (s *organizationService) CreateOrganization(entity *Organization) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *organizationService) GetOrganizationByID(id uint) (*Organization, error) {
	return s.repo.GetByID(id)
}

func (s *organizationService) UpdateOrganization(entity *Organization) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *organizationService) DeleteOrganization(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *organizationService) ListOrganizations(filter SearchFilter) ([]*Organization, error) {
	return s.repo.List(filter)
}
