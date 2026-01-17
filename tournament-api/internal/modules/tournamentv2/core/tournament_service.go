package core

type tournamentV2Service struct {
	repo TournamentV2Repository
}

func NewTournamentV2Service(repo TournamentV2Repository) TournamentV2Service {
	return &tournamentV2Service{repo: repo}
}

func (s *tournamentV2Service) CreateTournament(entity *TournamentV2) error {
	return s.repo.Create(entity)
}

func (s *tournamentV2Service) GetTournamentByID(id uint) (*TournamentV2, error) {
	return s.repo.GetByID(id)
}

func (s *tournamentV2Service) UpdateTournament(entity *TournamentV2) error {
	return s.repo.Update(entity)
}

func (s *tournamentV2Service) DeleteTournament(id uint) error {
	return s.repo.Delete(id)
}

func (s *tournamentV2Service) ListTournaments(filter SearchFilter) ([]*TournamentV2, error) {
	return s.repo.List(filter)
}
