// @kthulu:service:task
package core

type taskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(entity *Task) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *taskService) GetTaskByID(id uint) (*Task, error) {
	return s.repo.GetByID(id)
}

func (s *taskService) UpdateTask(entity *Task) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *taskService) DeleteTask(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *taskService) ListTasks(filter SearchFilter) ([]*Task, error) {
	return s.repo.List(filter)
}
