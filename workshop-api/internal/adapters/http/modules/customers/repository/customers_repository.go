// @kthulu:repository:customers
package repository

import (
        "gorm.io/gorm"

        "github.com/example/workshop-api/internal/adapters/http/modules/customers/domain"
)

type CustomersRepository struct {
        db *gorm.DB
}

func NewCustomersRepository(db *gorm.DB) domain.CustomersRepository {
        return &CustomersRepository{db: db}
}

func (r *CustomersRepository) Create(entity *domain.Customers) error {
        return r.db.Create(entity).Error
}

func (r *CustomersRepository) GetByID(id uint) (*domain.Customers, error) {
        var entity domain.Customers
        err := r.db.First(&entity, id).Error
        return &entity, err
}

func (r *CustomersRepository) Update(entity *domain.Customers) error {
        return r.db.Save(entity).Error
}

func (r *CustomersRepository) Delete(id uint) error {
        return r.db.Delete(&domain.Customers{}, id).Error
}

func (r *CustomersRepository) List(filter domain.SearchFilter) ([]*domain.Customers, error) {
        var entities []*domain.Customers
        query := r.db.Model(&domain.Customers{})

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
