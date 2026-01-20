// @kthulu:service:contact
package core

type contactService struct {
	repo ContactRepository
}

func NewContactService(repo ContactRepository) ContactService {
	return &contactService{repo: repo}
}

func (s *contactService) CreateContact(entity *Contact) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *contactService) GetContactByID(id uint) (*Contact, error) {
	return s.repo.GetByID(id)
}

func (s *contactService) UpdateContact(entity *Contact) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *contactService) DeleteContact(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *contactService) ListContacts(filter SearchFilter) ([]*Contact, error) {
	return s.repo.List(filter)
}
