// @kthulu:service:booking
package service

import (
	"airbnb-clone/internal/modules/booking/domain"
)

type BookingService struct {
	repo domain.BookingRepository
}

func NewBookingService(repo domain.BookingRepository) domain.BookingService {
	return &BookingService{repo: repo}
}

func (s *BookingService) CreateBooking(entity *domain.Booking) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *BookingService) GetBookingByID(id uint) (*domain.Booking, error) {
	return s.repo.GetByID(id)
}

func (s *BookingService) UpdateBooking(entity *domain.Booking) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *BookingService) DeleteBooking(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *BookingService) ListBookings(filter domain.SearchFilter) ([]*domain.Booking, error) {
	return s.repo.List(filter)
}
