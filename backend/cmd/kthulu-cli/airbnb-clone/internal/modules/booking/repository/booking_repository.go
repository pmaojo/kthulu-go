// @kthulu:repository:booking
package repository

import (
	"gorm.io/gorm"
	"airbnb-clone/internal/modules/booking/domain"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) domain.BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(entity *domain.Booking) error {
	return r.db.Create(entity).Error
}

func (r *BookingRepository) GetByID(id uint) (*domain.Booking, error) {
	var entity domain.Booking
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *BookingRepository) Update(entity *domain.Booking) error {
	return r.db.Save(entity).Error
}

func (r *BookingRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Booking{}, id).Error
}

func (r *BookingRepository) List(filter domain.SearchFilter) ([]*domain.Booking, error) {
	var entities []*domain.Booking
	query := r.db.Model(&domain.Booking{})

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
