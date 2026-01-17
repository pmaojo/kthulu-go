// @kthulu:service:tournament
package core

type tournamentService struct {
	repo TournamentRepository
}

func NewTournamentService(repo TournamentRepository) TournamentService {
	return &tournamentService{repo: repo}
}

func (s *tournamentService) CreateTournament(entity *Tournament) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *tournamentService) GetTournamentByID(id uint) (*Tournament, error) {
	return s.repo.GetByID(id)
}

func (s *tournamentService) UpdateTournament(entity *Tournament) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *tournamentService) DeleteTournament(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *tournamentService) ListTournaments(filter SearchFilter) ([]*Tournament, error) {
	return s.repo.List(filter)
}
