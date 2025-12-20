// @kthulu:service:orders
package service

import (
        "shop/internal/adapters/http/modules/orders/domain"
)

type OrdersService struct {
        repo domain.OrdersRepository
}

func NewOrdersService(repo domain.OrdersRepository) domain.OrdersService {
        return &OrdersService{repo: repo}
}

func (s *OrdersService) CreateOrders(entity *domain.Orders) error {
        // Add business logic here
        return s.repo.Create(entity)
}

func (s *OrdersService) GetOrdersByID(id uint) (*domain.Orders, error) {
        return s.repo.GetByID(id)
}

func (s *OrdersService) UpdateOrders(entity *domain.Orders) error {
        // Add business logic here
        return s.repo.Update(entity)
}

func (s *OrdersService) DeleteOrders(id uint) error {
        // Add business logic here
        return s.repo.Delete(id)
}

func (s *OrdersService) ListOrderss(filter domain.SearchFilter) ([]*domain.Orders, error) {
        return s.repo.List(filter)
}
