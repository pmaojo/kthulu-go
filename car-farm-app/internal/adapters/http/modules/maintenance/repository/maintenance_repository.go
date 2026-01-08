// @kthulu:repository:maintenance
package repository

import (
        "gorm.io/gorm"

        "car-farm-app/internal/adapters/http/modules/maintenance/domain"
)

type MaintenanceRepository struct {
        db *gorm.DB
}

func NewMaintenanceRepository(db *gorm.DB) domain.MaintenanceRepository {
        return &MaintenanceRepository{db: db}
}

func (r *MaintenanceRepository) Create(entity *domain.Maintenance) error {
        return r.db.Create(entity).Error
}

func (r *MaintenanceRepository) GetByID(id uint) (*domain.Maintenance, error) {
        var entity domain.Maintenance
        err := r.db.First(&entity, id).Error
        return &entity, err
}

func (r *MaintenanceRepository) Update(entity *domain.Maintenance) error {
        return r.db.Save(entity).Error
}

func (r *MaintenanceRepository) Delete(id uint) error {
        return r.db.Delete(&domain.Maintenance{}, id).Error
}

func (r *MaintenanceRepository) List(filter domain.SearchFilter) ([]*domain.Maintenance, error) {
        var entities []*domain.Maintenance
        query := r.db.Model(&domain.Maintenance{})

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
