// @kthulu:repository:payments
package repository

import (
        "gorm.io/gorm"

        "shop/internal/adapters/http/modules/payments/domain"
)

type PaymentsRepository struct {
        db *gorm.DB
}

func NewPaymentsRepository(db *gorm.DB) domain.PaymentsRepository {
        return &PaymentsRepository{db: db}
}

func (r *PaymentsRepository) Create(entity *domain.Payments) error {
        return r.db.Create(entity).Error
}

func (r *PaymentsRepository) GetByID(id uint) (*domain.Payments, error) {
        var entity domain.Payments
        err := r.db.First(&entity, id).Error
        return &entity, err
}

func (r *PaymentsRepository) Update(entity *domain.Payments) error {
        return r.db.Save(entity).Error
}

func (r *PaymentsRepository) Delete(id uint) error {
        return r.db.Delete(&domain.Payments{}, id).Error
}

func (r *PaymentsRepository) List(filter domain.SearchFilter) ([]*domain.Payments, error) {
        var entities []*domain.Payments
        query := r.db.Model(&domain.Payments{})

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
