// @kthulu:service:payments
package service

import (
        "shop/internal/adapters/http/modules/payments/domain"
)

type PaymentsService struct {
        repo domain.PaymentsRepository
}

func NewPaymentsService(repo domain.PaymentsRepository) domain.PaymentsService {
        return &PaymentsService{repo: repo}
}

func (s *PaymentsService) CreatePayments(entity *domain.Payments) error {
        // Add business logic here
        return s.repo.Create(entity)
}

func (s *PaymentsService) GetPaymentsByID(id uint) (*domain.Payments, error) {
        return s.repo.GetByID(id)
}

func (s *PaymentsService) UpdatePayments(entity *domain.Payments) error {
        // Add business logic here
        return s.repo.Update(entity)
}

func (s *PaymentsService) DeletePayments(id uint) error {
        // Add business logic here
        return s.repo.Delete(id)
}

func (s *PaymentsService) ListPaymentss(filter domain.SearchFilter) ([]*domain.Payments, error) {
        return s.repo.List(filter)
}
