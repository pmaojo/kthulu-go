// @kthulu:repository:services
package repository

import (
        "gorm.io/gorm"

        "github.com/example/workshop-api/internal/adapters/http/modules/services/domain"
)

type ServicesRepository struct {
        db *gorm.DB
}

func NewServicesRepository(db *gorm.DB) domain.ServicesRepository {
        return &ServicesRepository{db: db}
}

func (r *ServicesRepository) Create(entity *domain.Services) error {
        return r.db.Create(entity).Error
}

func (r *ServicesRepository) GetByID(id uint) (*domain.Services, error) {
        var entity domain.Services
        err := r.db.First(&entity, id).Error
        return &entity, err
}

func (r *ServicesRepository) Update(entity *domain.Services) error {
        return r.db.Save(entity).Error
}

func (r *ServicesRepository) Delete(id uint) error {
        return r.db.Delete(&domain.Services{}, id).Error
}

func (r *ServicesRepository) List(filter domain.SearchFilter) ([]*domain.Services, error) {
        var entities []*domain.Services
        query := r.db.Model(&domain.Services{})

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
