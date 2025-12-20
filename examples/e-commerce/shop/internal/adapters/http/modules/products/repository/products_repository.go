// @kthulu:repository:products
package repository

import (
        "gorm.io/gorm"

        "shop/internal/adapters/http/modules/products/domain"
)

type ProductsRepository struct {
        db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) domain.ProductsRepository {
        return &ProductsRepository{db: db}
}

func (r *ProductsRepository) Create(entity *domain.Products) error {
        return r.db.Create(entity).Error
}

func (r *ProductsRepository) GetByID(id uint) (*domain.Products, error) {
        var entity domain.Products
        err := r.db.First(&entity, id).Error
        return &entity, err
}

func (r *ProductsRepository) Update(entity *domain.Products) error {
        return r.db.Save(entity).Error
}

func (r *ProductsRepository) Delete(id uint) error {
        return r.db.Delete(&domain.Products{}, id).Error
}

func (r *ProductsRepository) List(filter domain.SearchFilter) ([]*domain.Products, error) {
        var entities []*domain.Products
        query := r.db.Model(&domain.Products{})

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
