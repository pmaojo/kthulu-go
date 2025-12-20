// @kthulu:service:products
package service

import (
        "fmt"
        "shop/internal/adapters/http/modules/products/domain"
)

type ProductsService struct {
        repo domain.ProductsRepository
}

func NewProductsService(repo domain.ProductsRepository) domain.ProductsService {
        return &ProductsService{repo: repo}
}

func (s *ProductsService) CreateProducts(entity *domain.Products) error {
        // Add business logic here
        if entity.Price < 0 {
                return fmt.Errorf("price cannot be negative")
        }
        return s.repo.Create(entity)
}

func (s *ProductsService) GetProductsByID(id uint) (*domain.Products, error) {
        return s.repo.GetByID(id)
}

func (s *ProductsService) UpdateProducts(entity *domain.Products) error {
        // Add business logic here
        return s.repo.Update(entity)
}

func (s *ProductsService) DeleteProducts(id uint) error {
        // Add business logic here
        return s.repo.Delete(id)
}

func (s *ProductsService) ListProductss(filter domain.SearchFilter) ([]*domain.Products, error) {
        return s.repo.List(filter)
}
