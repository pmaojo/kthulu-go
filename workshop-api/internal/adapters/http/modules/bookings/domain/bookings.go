// @kthulu:domain:bookings
package domain

import "time"

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Bookings represents a bookings entity
type Bookings struct {
        ID        uint      `json:"id" gorm:"primaryKey"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`

        Car_id int `json:"car_id" gorm:"car_id"`
        Service_id int `json:"service_id" gorm:"service_id"`
        Booking_date time.Time `json:"booking_date" gorm:"booking_date"`
        Status string `json:"status" gorm:"status"`

}

// TableName overrides the table name used by User to `bookingss`
func (Bookings) TableName() string {
	return "bookingss"
}

// BookingsRepository defines the repository interface
type BookingsRepository interface {
        Create(entity *Bookings) error
        GetByID(id uint) (*Bookings, error)
        Update(entity *Bookings) error
        Delete(id uint) error
        List(filter SearchFilter) ([]*Bookings, error)
}

// BookingsService defines the service interface
type BookingsService interface {
        CreateBookings(entity *Bookings) error
        GetBookingsByID(id uint) (*Bookings, error)
        UpdateBookings(entity *Bookings) error
        DeleteBookings(id uint) error
        ListBookingss(filter SearchFilter) ([]*Bookings, error)
}
