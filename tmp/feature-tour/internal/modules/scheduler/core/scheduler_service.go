// @kthulu:service:scheduler
package core

type schedulerService struct {
	repo SchedulerRepository
}

func NewSchedulerService(repo SchedulerRepository) SchedulerService {
	return &schedulerService{repo: repo}
}

func (s *schedulerService) CreateScheduler(entity *Scheduler) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *schedulerService) GetSchedulerByID(id uint) (*Scheduler, error) {
	return s.repo.GetByID(id)
}

func (s *schedulerService) UpdateScheduler(entity *Scheduler) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *schedulerService) DeleteScheduler(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *schedulerService) ListSchedulers(filter SearchFilter) ([]*Scheduler, error) {
	return s.repo.List(filter)
}
