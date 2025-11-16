// @kthulu:repository:product
package repository

import (
	"gorm.io/gorm"
	"my-kthulu-app/internal/adapters/http/modules/product/domain"
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

func (r *ProductRepository) List() ([]*domain.Product, error) {
	var entities []*domain.Product
	err := r.db.Find(&entities).Error
	return entities, err
}