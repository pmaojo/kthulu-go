// @kthulu:service:contact
package service

import (
	"my-kthulu-app/internal/adapters/http/modules/contact/domain"
)

type ContactService struct {
	repo domain.ContactRepository
}

func NewContactService(repo domain.ContactRepository) domain.ContactService {
	return &ContactService{repo: repo}
}

func (s *ContactService) CreateContact(entity *domain.Contact) error {
	return s.repo.Create(entity)
}

func (s *ContactService) GetContactByID(id uint) (*domain.Contact, error) {
	return s.repo.GetByID(id)
}

func (s *ContactService) UpdateContact(entity *domain.Contact) error {
	return s.repo.Update(entity)
}

func (s *ContactService) DeleteContact(id uint) error {
	return s.repo.Delete(id)
}

func (s *ContactService) ListContacts() ([]*domain.Contact, error) {
	return s.repo.List()
}