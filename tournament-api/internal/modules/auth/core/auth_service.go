// @kthulu:service:auth
package core

type authService struct {
	repo AuthRepository
}

func NewAuthService(repo AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) CreateAuth(entity *Auth) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *authService) GetAuthByID(id uint) (*Auth, error) {
	return s.repo.GetByID(id)
}

func (s *authService) UpdateAuth(entity *Auth) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *authService) DeleteAuth(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *authService) ListAuths(filter SearchFilter) ([]*Auth, error) {
	return s.repo.List(filter)
}
