// @kthulu:service:user
package core

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(entity *User) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *UserService) GetUserByID(id uint) (*User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) UpdateUser(entity *User) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *UserService) DeleteUser(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *UserService) ListUsers(filter SearchFilter) ([]*User, error) {
	return s.repo.List(filter)
}
