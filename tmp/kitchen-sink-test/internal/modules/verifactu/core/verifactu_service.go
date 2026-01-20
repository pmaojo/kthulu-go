// @kthulu:service:verifactu
package core

type verifactuService struct {
	repo VerifactuRepository
}

func NewVerifactuService(repo VerifactuRepository) VerifactuService {
	return &verifactuService{repo: repo}
}

func (s *verifactuService) CreateVerifactu(entity *Verifactu) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *verifactuService) GetVerifactuByID(id uint) (*Verifactu, error) {
	return s.repo.GetByID(id)
}

func (s *verifactuService) UpdateVerifactu(entity *Verifactu) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *verifactuService) DeleteVerifactu(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *verifactuService) ListVerifactus(filter SearchFilter) ([]*Verifactu, error) {
	return s.repo.List(filter)
}
