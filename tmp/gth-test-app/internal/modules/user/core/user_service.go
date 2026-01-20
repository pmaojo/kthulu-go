// @kthulu:service:user
package core

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(entity *User) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *userService) GetUserByID(id uint) (*User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) UpdateUser(entity *User) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *userService) DeleteUser(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *userService) ListUsers(filter SearchFilter) ([]*User, error) {
	return s.repo.List(filter)
}
