// @kthulu:store:product
package store

import (
	"gorm.io/gorm"
	"gth-test-app/internal/modules/product/core"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) core.ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(entity *core.Product) error {
	return r.db.Create(entity).Error
}

func (r *ProductRepository) GetByID(id uint) (*core.Product, error) {
	var entity core.Product
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *ProductRepository) Update(entity *core.Product) error {
	return r.db.Save(entity).Error
}

func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&core.Product{}, id).Error
}

func (r *ProductRepository) List(filter core.SearchFilter) ([]*core.Product, error) {
	var entities []*core.Product
	query := r.db.Model(&core.Product{})

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
