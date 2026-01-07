// @kthulu:service:bookings
package service

import (
        "github.com/example/workshop-api/internal/adapters/http/modules/bookings/domain"
)

type BookingsService struct {
        repo domain.BookingsRepository
}

func NewBookingsService(repo domain.BookingsRepository) domain.BookingsService {
        return &BookingsService{repo: repo}
}

func (s *BookingsService) CreateBookings(entity *domain.Bookings) error {
        // Add business logic here
        return s.repo.Create(entity)
}

func (s *BookingsService) GetBookingsByID(id uint) (*domain.Bookings, error) {
        return s.repo.GetByID(id)
}

func (s *BookingsService) UpdateBookings(entity *domain.Bookings) error {
        // Add business logic here
        return s.repo.Update(entity)
}

func (s *BookingsService) DeleteBookings(id uint) error {
        // Add business logic here
        return s.repo.Delete(id)
}

func (s *BookingsService) ListBookingss(filter domain.SearchFilter) ([]*domain.Bookings, error) {
        return s.repo.List(filter)
}
