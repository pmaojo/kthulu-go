// @kthulu:store:contact
package store

import (
	"gorm.io/gorm"
	"feature-tour/internal/modules/contact/core"
)

type ContactRepository struct {
	db *gorm.DB
}

func NewContactRepository(db *gorm.DB) core.ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Create(entity *core.Contact) error {
	return r.db.Create(entity).Error
}

func (r *ContactRepository) GetByID(id uint) (*core.Contact, error) {
	var entity core.Contact
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *ContactRepository) Update(entity *core.Contact) error {
	return r.db.Save(entity).Error
}

func (r *ContactRepository) Delete(id uint) error {
	return r.db.Delete(&core.Contact{}, id).Error
}

func (r *ContactRepository) List(filter core.SearchFilter) ([]*core.Contact, error) {
	var entities []*core.Contact
	query := r.db.Model(&core.Contact{})

	if filter.Query != "" {
		// Basic search implementation
		// query = query.Where("name LIKE ?", "%"+filter.Query+"%")
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Find(&entities).Error
	return entities, err
}
