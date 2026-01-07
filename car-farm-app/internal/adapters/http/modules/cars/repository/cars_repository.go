// @kthulu:repository:cars
package repository

import (
        "gorm.io/gorm"

        "car-farm-app/internal/adapters/http/modules/cars/domain"
)

type CarsRepository struct {
        db *gorm.DB
}

func NewCarsRepository(db *gorm.DB) domain.CarsRepository {
        return &CarsRepository{db: db}
}

func (r *CarsRepository) Create(entity *domain.Cars) error {
        return r.db.Create(entity).Error
}

func (r *CarsRepository) GetByID(id uint) (*domain.Cars, error) {
        var entity domain.Cars
        err := r.db.First(&entity, id).Error
        return &entity, err
}

func (r *CarsRepository) Update(entity *domain.Cars) error {
        return r.db.Save(entity).Error
}

func (r *CarsRepository) Delete(id uint) error {
        return r.db.Delete(&domain.Cars{}, id).Error
}

func (r *CarsRepository) List(filter domain.SearchFilter) ([]*domain.Cars, error) {
        var entities []*domain.Cars
        query := r.db.Model(&domain.Cars{})

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
