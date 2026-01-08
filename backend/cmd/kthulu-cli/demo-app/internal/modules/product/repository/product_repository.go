// @kthulu:repository:product
package repository

import (
	"gorm.io/gorm"
	"demo-app/internal/modules/product/domain"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(entity *domain.Product) error {
	return r.db.Create(entity).Error
}

func (r *ProductRepository) GetByID(id uint) (*domain.Product, error) {
	var entity domain.Product
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *ProductRepository) Update(entity *domain.Product) error {
	return r.db.Save(entity).Error
}

func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Product{}, id).Error
}

func (r *ProductRepository) List(filter domain.SearchFilter) ([]*domain.Product, error) {
	var entities []*domain.Product
	query := r.db.Model(&domain.Product{})

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
