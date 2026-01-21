// @kthulu:store:mail
package store

import (
	"gorm.io/gorm"
	"feature-tour/internal/modules/mail/core"
)

type MailRepository struct {
	db *gorm.DB
}

func NewMailRepository(db *gorm.DB) core.MailRepository {
	return &MailRepository{db: db}
}

func (r *MailRepository) Create(entity *core.Mail) error {
	return r.db.Create(entity).Error
}

func (r *MailRepository) GetByID(id uint) (*core.Mail, error) {
	var entity core.Mail
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *MailRepository) Update(entity *core.Mail) error {
	return r.db.Save(entity).Error
}

func (r *MailRepository) Delete(id uint) error {
	return r.db.Delete(&core.Mail{}, id).Error
}

func (r *MailRepository) List(filter core.SearchFilter) ([]*core.Mail, error) {
	var entities []*core.Mail
	query := r.db.Model(&core.Mail{})

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
