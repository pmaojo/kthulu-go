// @kthulu:domain:booking
package domain

import (
	"time"
	
)

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Booking represents a booking entity
type Booking struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	PropertyID string `json:"property_i_d" gorm:"property_i_d"`
	GuestID string `json:"guest_i_d" gorm:"guest_i_d"`
	CheckIn string `json:"check_in" gorm:"check_in"`
	CheckOut string `json:"check_out" gorm:"check_out"`
	TotalPrice string `json:"total_price" gorm:"total_price"`
	Status string `json:"status" gorm:"status"`
}

// TableName overrides the table name used by Booking to `Bookings`
func (Booking) TableName() string {
	return "Bookings"
}

// BookingRepository defines the repository interface
type BookingRepository interface {
	Create(entity *Booking) error
	GetByID(id uint) (*Booking, error)
	Update(entity *Booking) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Booking, error)
}

// BookingService defines the service interface  
type BookingService interface {
	CreateBooking(entity *Booking) error
	GetBookingByID(id uint) (*Booking, error)
	UpdateBooking(entity *Booking) error
	DeleteBooking(id uint) error
	ListBookings(filter SearchFilter) ([]*Booking, error)
}
