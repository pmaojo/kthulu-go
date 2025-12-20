// @kthulu:repository:orders
package repository

import (
        "gorm.io/gorm"

        "shop/internal/adapters/http/modules/orders/domain"
)

type OrdersRepository struct {
        db *gorm.DB
}

func NewOrdersRepository(db *gorm.DB) domain.OrdersRepository {
        return &OrdersRepository{db: db}
}

func (r *OrdersRepository) Create(entity *domain.Orders) error {
        return r.db.Create(entity).Error
}

func (r *OrdersRepository) GetByID(id uint) (*domain.Orders, error) {
        var entity domain.Orders
        err := r.db.First(&entity, id).Error
        return &entity, err
}

func (r *OrdersRepository) Update(entity *domain.Orders) error {
        return r.db.Save(entity).Error
}

func (r *OrdersRepository) Delete(id uint) error {
        return r.db.Delete(&domain.Orders{}, id).Error
}

func (r *OrdersRepository) List(filter domain.SearchFilter) ([]*domain.Orders, error) {
        var entities []*domain.Orders
        query := r.db.Model(&domain.Orders{})

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
