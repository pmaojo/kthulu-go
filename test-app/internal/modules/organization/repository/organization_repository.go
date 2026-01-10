// @kthulu:repository:organization
package repository

import (
	"gorm.io/gorm"
	"test-app/internal/modules/organization/domain"
)

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) domain.OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(entity *domain.Organization) error {
	return r.db.Create(entity).Error
}

func (r *OrganizationRepository) GetByID(id uint) (*domain.Organization, error) {
	var entity domain.Organization
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *OrganizationRepository) Update(entity *domain.Organization) error {
	return r.db.Save(entity).Error
}

func (r *OrganizationRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Organization{}, id).Error
}

func (r *OrganizationRepository) List(filter domain.SearchFilter) ([]*domain.Organization, error) {
	var entities []*domain.Organization
	query := r.db.Model(&domain.Organization{})

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
