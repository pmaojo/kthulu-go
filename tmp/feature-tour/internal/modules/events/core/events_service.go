// @kthulu:service:events
package core

type eventService struct {
	repo EventRepository
}

func NewEventService(repo EventRepository) EventService {
	return &eventService{repo: repo}
}

func (s *eventService) CreateEvent(entity *Event) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *eventService) GetEventByID(id uint) (*Event, error) {
	return s.repo.GetByID(id)
}

func (s *eventService) UpdateEvent(entity *Event) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *eventService) DeleteEvent(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *eventService) ListEvents(filter SearchFilter) ([]*Event, error) {
	return s.repo.List(filter)
}
