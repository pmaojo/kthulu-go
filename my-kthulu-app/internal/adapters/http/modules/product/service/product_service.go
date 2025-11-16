// @kthulu:service:product
package service

import (
	"my-kthulu-app/internal/adapters/http/modules/product/domain"
)

type ProductService struct {
	repo domain.ProductRepository
}

func NewProductService(repo domain.ProductRepository) domain.ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(entity *domain.Product) error {
	return s.repo.Create(entity)
}

func (s *ProductService) GetProductByID(id uint) (*domain.Product, error) {
	return s.repo.GetByID(id)
}

func (s *ProductService) UpdateProduct(entity *domain.Product) error {
	return s.repo.Update(entity)
}

func (s *ProductService) DeleteProduct(id uint) error {
	return s.repo.Delete(id)
}

func (s *ProductService) ListProducts() ([]*domain.Product, error) {
	return s.repo.List()
}