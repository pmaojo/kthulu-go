// @kthulu:repository:property
package repository

import (
	"gorm.io/gorm"
	"airbnb-v2/internal/modules/property/domain"
)

type PropertyRepository struct {
	db *gorm.DB
}

func NewPropertyRepository(db *gorm.DB) domain.PropertyRepository {
	return &PropertyRepository{db: db}
}

func (r *PropertyRepository) Create(entity *domain.Property) error {
	return r.db.Create(entity).Error
}

func (r *PropertyRepository) GetByID(id uint) (*domain.Property, error) {
	var entity domain.Property
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *PropertyRepository) Update(entity *domain.Property) error {
	return r.db.Save(entity).Error
}

func (r *PropertyRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Property{}, id).Error
}

func (r *PropertyRepository) List(filter domain.SearchFilter) ([]*domain.Property, error) {
	var entities []*domain.Property
	query := r.db.Model(&domain.Property{})

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
