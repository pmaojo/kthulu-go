// @kthulu:service:debugmod
package core

type debugmodService struct {
	repo DebugmodRepository
}

func NewDebugmodService(repo DebugmodRepository) DebugmodService {
	return &debugmodService{repo: repo}
}

func (s *debugmodService) CreateDebugmod(entity *Debugmod) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *debugmodService) GetDebugmodByID(id uint) (*Debugmod, error) {
	return s.repo.GetByID(id)
}

func (s *debugmodService) UpdateDebugmod(entity *Debugmod) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *debugmodService) DeleteDebugmod(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *debugmodService) ListDebugmods(filter SearchFilter) ([]*Debugmod, error) {
	return s.repo.List(filter)
}
