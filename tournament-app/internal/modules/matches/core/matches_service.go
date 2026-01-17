// @kthulu:service:matches
package core

type matchService struct {
	repo MatchRepository
}

func NewMatchService(repo MatchRepository) MatchService {
	return &matchService{repo: repo}
}

func (s *matchService) CreateMatch(entity *Match) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *matchService) GetMatchByID(id uint) (*Match, error) {
	return s.repo.GetByID(id)
}

func (s *matchService) UpdateMatch(entity *Match) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *matchService) DeleteMatch(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *matchService) ListMatches(filter SearchFilter) ([]*Match, error) {
	return s.repo.List(filter)
}
