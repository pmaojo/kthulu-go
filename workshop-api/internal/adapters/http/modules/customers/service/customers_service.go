// @kthulu:service:customers
package service

import (
        "github.com/example/workshop-api/internal/adapters/http/modules/customers/domain"
)

type CustomersService struct {
        repo domain.CustomersRepository
}

func NewCustomersService(repo domain.CustomersRepository) domain.CustomersService {
        return &CustomersService{repo: repo}
}

func (s *CustomersService) CreateCustomers(entity *domain.Customers) error {
        // Add business logic here
        return s.repo.Create(entity)
}

func (s *CustomersService) GetCustomersByID(id uint) (*domain.Customers, error) {
        return s.repo.GetByID(id)
}

func (s *CustomersService) UpdateCustomers(entity *domain.Customers) error {
        // Add business logic here
        return s.repo.Update(entity)
}

func (s *CustomersService) DeleteCustomers(id uint) error {
        // Add business logic here
        return s.repo.Delete(id)
}

func (s *CustomersService) ListCustomerss(filter domain.SearchFilter) ([]*domain.Customers, error) {
        return s.repo.List(filter)
}
