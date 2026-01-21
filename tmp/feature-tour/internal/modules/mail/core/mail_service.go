// @kthulu:service:mail
package core

type mailService struct {
	repo MailRepository
}

func NewMailService(repo MailRepository) MailService {
	return &mailService{repo: repo}
}

func (s *mailService) CreateMail(entity *Mail) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *mailService) GetMailByID(id uint) (*Mail, error) {
	return s.repo.GetByID(id)
}

func (s *mailService) UpdateMail(entity *Mail) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *mailService) DeleteMail(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *mailService) ListMails(filter SearchFilter) ([]*Mail, error) {
	return s.repo.List(filter)
}
