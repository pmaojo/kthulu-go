// @kthulu:store:organization
package store

import (
	"gorm.io/gorm"
	"kitchen-sink-test/internal/modules/organization/core"
)

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) core.OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(entity *core.Organization) error {
	return r.db.Create(entity).Error
}

func (r *OrganizationRepository) GetByID(id uint) (*core.Organization, error) {
	var entity core.Organization
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *OrganizationRepository) Update(entity *core.Organization) error {
	return r.db.Save(entity).Error
}

func (r *OrganizationRepository) Delete(id uint) error {
	return r.db.Delete(&core.Organization{}, id).Error
}

func (r *OrganizationRepository) List(filter core.SearchFilter) ([]*core.Organization, error) {
	var entities []*core.Organization
	query := r.db.Model(&core.Organization{})

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
