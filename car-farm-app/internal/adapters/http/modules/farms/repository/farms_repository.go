// @kthulu:repository:farms
package repository

import (
        "gorm.io/gorm"

        "car-farm-app/internal/adapters/http/modules/farms/domain"
)

type FarmsRepository struct {
        db *gorm.DB
}

func NewFarmsRepository(db *gorm.DB) domain.FarmsRepository {
        return &FarmsRepository{db: db}
}

func (r *FarmsRepository) Create(entity *domain.Farms) error {
        return r.db.Create(entity).Error
}

func (r *FarmsRepository) GetByID(id uint) (*domain.Farms, error) {
        var entity domain.Farms
        err := r.db.First(&entity, id).Error
        return &entity, err
}

func (r *FarmsRepository) Update(entity *domain.Farms) error {
        return r.db.Save(entity).Error
}

func (r *FarmsRepository) Delete(id uint) error {
        return r.db.Delete(&domain.Farms{}, id).Error
}

func (r *FarmsRepository) List(filter domain.SearchFilter) ([]*domain.Farms, error) {
        var entities []*domain.Farms
        query := r.db.Model(&domain.Farms{})

		if filter.Query != "" {
			// Basic search implementation
			// Note: Adjust fields based on your actual model
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
