// @kthulu:service:participant
package core

type participantService struct {
	repo ParticipantRepository
}

func NewParticipantService(repo ParticipantRepository) ParticipantService {
	return &participantService{repo: repo}
}

func (s *participantService) CreateParticipant(entity *Participant) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *participantService) GetParticipantByID(id uint) (*Participant, error) {
	return s.repo.GetByID(id)
}

func (s *participantService) UpdateParticipant(entity *Participant) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *participantService) DeleteParticipant(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *participantService) ListParticipants(filter SearchFilter) ([]*Participant, error) {
	return s.repo.List(filter)
}
