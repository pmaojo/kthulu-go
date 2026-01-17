// @kthulu:service:todo
package core

type todoService struct {
	repo TodoRepository
}

func NewTodoService(repo TodoRepository) TodoService {
	return &todoService{repo: repo}
}

func (s *todoService) CreateTodo(entity *Todo) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *todoService) GetTodoByID(id uint) (*Todo, error) {
	return s.repo.GetByID(id)
}

func (s *todoService) UpdateTodo(entity *Todo) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *todoService) DeleteTodo(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *todoService) ListTodos(filter SearchFilter) ([]*Todo, error) {
	return s.repo.List(filter)
}
