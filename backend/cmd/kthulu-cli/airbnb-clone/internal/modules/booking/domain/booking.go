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

	PropertyID uint      `json:"property_id"`
	GuestID    uint      `json:"guest_id"`
	CheckIn    time.Time `json:"check_in"`
	CheckOut   time.Time `json:"check_out"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"` // pending, confirmed, cancelled
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
