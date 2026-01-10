// @kthulu:service:auth
package core

type AuthService struct {
	repo AuthRepository
}

func NewAuthService(repo AuthRepository) AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) CreateAuth(entity *Auth) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *AuthService) GetAuthByID(id uint) (*Auth, error) {
	return s.repo.GetByID(id)
}

func (s *AuthService) UpdateAuth(entity *Auth) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *AuthService) DeleteAuth(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *AuthService) ListAuths(filter SearchFilter) ([]*Auth, error) {
	return s.repo.List(filter)
}
