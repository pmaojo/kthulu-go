// @kthulu:service:calendar
package core

type calendarService struct {
	repo CalendarRepository
}

func NewCalendarService(repo CalendarRepository) CalendarService {
	return &calendarService{repo: repo}
}

func (s *calendarService) CreateCalendar(entity *Calendar) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *calendarService) GetCalendarByID(id uint) (*Calendar, error) {
	return s.repo.GetByID(id)
}

func (s *calendarService) UpdateCalendar(entity *Calendar) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *calendarService) DeleteCalendar(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *calendarService) ListCalendars(filter SearchFilter) ([]*Calendar, error) {
	return s.repo.List(filter)
}
