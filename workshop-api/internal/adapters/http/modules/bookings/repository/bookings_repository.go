// @kthulu:repository:bookings
package repository

import (
        "gorm.io/gorm"

        "github.com/example/workshop-api/internal/adapters/http/modules/bookings/domain"
)

type BookingsRepository struct {
        db *gorm.DB
}

func NewBookingsRepository(db *gorm.DB) domain.BookingsRepository {
        return &BookingsRepository{db: db}
}

func (r *BookingsRepository) Create(entity *domain.Bookings) error {
        return r.db.Create(entity).Error
}

func (r *BookingsRepository) GetByID(id uint) (*domain.Bookings, error) {
        var entity domain.Bookings
        err := r.db.First(&entity, id).Error
        return &entity, err
}

func (r *BookingsRepository) Update(entity *domain.Bookings) error {
        return r.db.Save(entity).Error
}

func (r *BookingsRepository) Delete(id uint) error {
        return r.db.Delete(&domain.Bookings{}, id).Error
}

func (r *BookingsRepository) List(filter domain.SearchFilter) ([]*domain.Bookings, error) {
        var entities []*domain.Bookings
        query := r.db.Model(&domain.Bookings{})

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
