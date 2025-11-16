// @kthulu:service:product
package service

import (
	"test-project/internal/adapters/http/modules/product/domain"
)

type ProductService struct {
	repo domain.ProductRepository
}

func NewProductService(repo domain.ProductRepository) domain.ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(entity *domain.Product) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *ProductService) GetProductByID(id uint) (*domain.Product, error) {
	return s.repo.GetByID(id)
}

func (s *ProductService) UpdateProduct(entity *domain.Product) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *ProductService) DeleteProduct(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *ProductsService) ListProduct() ([]*domain.%!s(MISSING), error) {
	return s.repo.List()
}
