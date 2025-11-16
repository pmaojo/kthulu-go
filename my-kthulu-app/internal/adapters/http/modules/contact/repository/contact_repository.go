// @kthulu:repository:contact
package repository

import (
	"gorm.io/gorm"
	"my-kthulu-app/internal/adapters/http/modules/contact/domain"
)

type ContactRepository struct {
	db *gorm.DB
}

func NewContactRepository(db *gorm.DB) domain.ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Create(entity *domain.Contact) error {
	return r.db.Create(entity).Error
}

func (r *ContactRepository) GetByID(id uint) (*domain.Contact, error) {
	var entity domain.Contact
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *ContactRepository) Update(entity *domain.Contact) error {
	return r.db.Save(entity).Error
}

func (r *ContactRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Contact{}, id).Error
}

func (r *ContactRepository) List() ([]*domain.Contact, error) {
	var entities []*domain.Contact
	err := r.db.Find(&entities).Error
	return entities, err
}