// @kthulu:module:calendar
package repository

import (
	"context"
	"time"

	"github.com/kthulu/kthulu-go/backend/internal/domain"
)

// CalendarRepository defines the interface for calendar data access
type CalendarRepository interface {
	// Calendar operations
	CreateCalendar(ctx context.Context, calendar *domain.Calendar) error
	GetCalendar(ctx context.Context, id uint) (*domain.Calendar, error)
	ListCalendars(ctx context.Context, ownerID uint, page, pageSize int) ([]domain.Calendar, int, error)
	UpdateCalendar(ctx context.Context, calendar *domain.Calendar) error
	DeleteCalendar(ctx context.Context, id uint) error

	// Event operations
	CreateEvent(ctx context.Context, event *domain.Event) error
	GetEvent(ctx context.Context, id uint) (*domain.Event, error)
	ListEvents(ctx context.Context, calendarID uint, page, pageSize int) ([]domain.Event, int, error)
	GetEventsByTimeRange(ctx context.Context, calendarID uint, start, end time.Time) ([]domain.Event, error)
	UpdateEvent(ctx context.Context, event *domain.Event) error
	DeleteEvent(ctx context.Context, id uint) error

	// Attendee operations
	CreateAttendee(ctx context.Context, attendee *domain.Attendee) error
	GetAttendees(ctx context.Context, eventID uint) ([]domain.Attendee, error)
	UpdateAttendee(ctx context.Context, attendee *domain.Attendee) error
	DeleteAttendee(ctx context.Context, id uint) error

	// Reminder operations
	CreateReminder(ctx context.Context, reminder *domain.Reminder) error
	GetReminders(ctx context.Context, eventID uint) ([]domain.Reminder, error)
	GetPendingReminders(ctx context.Context, before time.Time) ([]domain.Reminder, error)
	UpdateReminder(ctx context.Context, reminder *domain.Reminder) error
	DeleteReminder(ctx context.Context, id uint) error

	// Availability slot operations
	CreateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) error
	GetAvailabilitySlot(ctx context.Context, id uint) (*domain.AvailabilitySlot, error)
	ListAvailabilitySlots(ctx context.Context, calendarID uint, start, end time.Time) ([]domain.AvailabilitySlot, error)
	UpdateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) error
	DeleteAvailabilitySlot(ctx context.Context, id uint) error

	// Booking operations
	CreateBooking(ctx context.Context, booking *domain.Booking) error
	GetBooking(ctx context.Context, id uint) (*domain.Booking, error)
	ListBookings(ctx context.Context, calendarID uint, page, pageSize int) ([]domain.Booking, int, error)
	GetBookingsByTimeRange(ctx context.Context, calendarID uint, start, end time.Time) ([]domain.Booking, error)
	UpdateBooking(ctx context.Context, booking *domain.Booking) error
	DeleteBooking(ctx context.Context, id uint) error

	// Search and filtering
	SearchEvents(ctx context.Context, calendarID uint, query string, page, pageSize int) ([]domain.Event, int, error)
	GetEventsByStatus(ctx context.Context, calendarID uint, status domain.EventStatus) ([]domain.Event, error)
	GetEventsByType(ctx context.Context, calendarID uint, eventType domain.EventType) ([]domain.Event, error)
}
