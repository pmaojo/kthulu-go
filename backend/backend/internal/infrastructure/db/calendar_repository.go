// @kthulu:module:calendar
package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// calendarRepository implements the CalendarRepository interface using GORM
type calendarRepository struct {
	db *gorm.DB
}

// NewCalendarRepository creates a new calendar repository
func NewCalendarRepository(db *gorm.DB) repository.CalendarRepository {
	return &calendarRepository{db: db}
}

// Calendar operations
func (r *calendarRepository) CreateCalendar(ctx context.Context, calendar *domain.Calendar) error {
	return r.db.WithContext(ctx).Create(calendar).Error
}

func (r *calendarRepository) GetCalendar(ctx context.Context, id uint) (*domain.Calendar, error) {
	var calendar domain.Calendar
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Preload("Events").
		First(&calendar, id).Error
	if err != nil {
		return nil, err
	}
	return &calendar, nil
}

func (r *calendarRepository) ListCalendars(ctx context.Context, ownerID uint, page, pageSize int) ([]domain.Calendar, int, error) {
	var calendars []domain.Calendar
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Calendar{}).
		Where("owner_id = ?", ownerID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("name ASC").
		Find(&calendars).Error

	return calendars, int(total), err
}

func (r *calendarRepository) UpdateCalendar(ctx context.Context, calendar *domain.Calendar) error {
	return r.db.WithContext(ctx).Save(calendar).Error
}

func (r *calendarRepository) DeleteCalendar(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Calendar{}, id).Error
}

// Event operations
func (r *calendarRepository) CreateEvent(ctx context.Context, event *domain.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *calendarRepository) GetEvent(ctx context.Context, id uint) (*domain.Event, error) {
	var event domain.Event
	err := r.db.WithContext(ctx).
		Preload("Calendar").
		Preload("CreatedBy").
		Preload("Attendees").
		Preload("Reminders").
		First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *calendarRepository) ListEvents(ctx context.Context, calendarID uint, page, pageSize int) ([]domain.Event, int, error) {
	var events []domain.Event
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Event{}).
		Where("calendar_id = ?", calendarID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("start_time ASC").
		Find(&events).Error

	return events, int(total), err
}

func (r *calendarRepository) GetEventsByTimeRange(ctx context.Context, calendarID uint, start, end time.Time) ([]domain.Event, error) {
	var events []domain.Event
	err := r.db.WithContext(ctx).
		Where("calendar_id = ? AND start_time < ? AND end_time > ?", calendarID, end, start).
		Order("start_time ASC").
		Find(&events).Error
	return events, err
}

func (r *calendarRepository) UpdateEvent(ctx context.Context, event *domain.Event) error {
	return r.db.WithContext(ctx).Save(event).Error
}

func (r *calendarRepository) DeleteEvent(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Event{}, id).Error
}

// Attendee operations
func (r *calendarRepository) CreateAttendee(ctx context.Context, attendee *domain.Attendee) error {
	return r.db.WithContext(ctx).Create(attendee).Error
}

func (r *calendarRepository) GetAttendees(ctx context.Context, eventID uint) ([]domain.Attendee, error) {
	var attendees []domain.Attendee
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("event_id = ?", eventID).
		Find(&attendees).Error
	return attendees, err
}

func (r *calendarRepository) UpdateAttendee(ctx context.Context, attendee *domain.Attendee) error {
	return r.db.WithContext(ctx).Save(attendee).Error
}

func (r *calendarRepository) DeleteAttendee(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Attendee{}, id).Error
}

// Reminder operations
func (r *calendarRepository) CreateReminder(ctx context.Context, reminder *domain.Reminder) error {
	return r.db.WithContext(ctx).Create(reminder).Error
}

func (r *calendarRepository) GetReminders(ctx context.Context, eventID uint) ([]domain.Reminder, error) {
	var reminders []domain.Reminder
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Find(&reminders).Error
	return reminders, err
}

func (r *calendarRepository) GetPendingReminders(ctx context.Context, before time.Time) ([]domain.Reminder, error) {
	var reminders []domain.Reminder
	err := r.db.WithContext(ctx).
		Preload("Event").
		Where("is_sent = ? AND created_at <= ?", false, before).
		Find(&reminders).Error
	return reminders, err
}

func (r *calendarRepository) UpdateReminder(ctx context.Context, reminder *domain.Reminder) error {
	return r.db.WithContext(ctx).Save(reminder).Error
}

func (r *calendarRepository) DeleteReminder(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Reminder{}, id).Error
}

// Availability slot operations
func (r *calendarRepository) CreateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) error {
	return r.db.WithContext(ctx).Create(slot).Error
}

func (r *calendarRepository) GetAvailabilitySlot(ctx context.Context, id uint) (*domain.AvailabilitySlot, error) {
	var slot domain.AvailabilitySlot
	err := r.db.WithContext(ctx).
		Preload("Calendar").
		First(&slot, id).Error
	if err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *calendarRepository) ListAvailabilitySlots(ctx context.Context, calendarID uint, start, end time.Time) ([]domain.AvailabilitySlot, error) {
	var slots []domain.AvailabilitySlot
	err := r.db.WithContext(ctx).
		Where("calendar_id = ? AND start_time >= ? AND end_time <= ?", calendarID, start, end).
		Order("start_time ASC").
		Find(&slots).Error
	return slots, err
}

func (r *calendarRepository) UpdateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) error {
	return r.db.WithContext(ctx).Save(slot).Error
}

func (r *calendarRepository) DeleteAvailabilitySlot(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.AvailabilitySlot{}, id).Error
}

// Booking operations
func (r *calendarRepository) CreateBooking(ctx context.Context, booking *domain.Booking) error {
	return r.db.WithContext(ctx).Create(booking).Error
}

func (r *calendarRepository) GetBooking(ctx context.Context, id uint) (*domain.Booking, error) {
	var booking domain.Booking
	err := r.db.WithContext(ctx).
		Preload("AvailabilitySlot").
		Preload("Event").
		Preload("BookedBy").
		First(&booking, id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *calendarRepository) ListBookings(ctx context.Context, calendarID uint, page, pageSize int) ([]domain.Booking, int, error) {
	var bookings []domain.Booking
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Booking{}).
		Joins("JOIN availability_slots ON bookings.availability_slot_id = availability_slots.id").
		Where("availability_slots.calendar_id = ?", calendarID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Preload("AvailabilitySlot").
		Preload("Event").
		Preload("BookedBy").
		Order("bookings.created_at DESC").
		Find(&bookings).Error

	return bookings, int(total), err
}

func (r *calendarRepository) GetBookingsByTimeRange(ctx context.Context, calendarID uint, start, end time.Time) ([]domain.Booking, error) {
	var bookings []domain.Booking
	err := r.db.WithContext(ctx).
		Preload("AvailabilitySlot").
		Preload("Event").
		Preload("BookedBy").
		Joins("JOIN availability_slots ON bookings.availability_slot_id = availability_slots.id").
		Where("availability_slots.calendar_id = ? AND availability_slots.start_time >= ? AND availability_slots.end_time <= ?", calendarID, start, end).
		Order("availability_slots.start_time ASC").
		Find(&bookings).Error
	return bookings, err
}

func (r *calendarRepository) UpdateBooking(ctx context.Context, booking *domain.Booking) error {
	return r.db.WithContext(ctx).Save(booking).Error
}

func (r *calendarRepository) DeleteBooking(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Booking{}, id).Error
}

// Search and filtering
func (r *calendarRepository) SearchEvents(ctx context.Context, calendarID uint, query string, page, pageSize int) ([]domain.Event, int, error) {
	var events []domain.Event
	var total int64

	searchQuery := r.db.WithContext(ctx).Model(&domain.Event{}).
		Where("calendar_id = ?", calendarID)

	if query != "" {
		searchPattern := "%" + query + "%"
		searchQuery = searchQuery.Where("title ILIKE ? OR description ILIKE ? OR location ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := searchQuery.Offset(offset).Limit(pageSize).
		Order("start_time ASC").
		Find(&events).Error

	return events, int(total), err
}

func (r *calendarRepository) GetEventsByStatus(ctx context.Context, calendarID uint, status domain.EventStatus) ([]domain.Event, error) {
	var events []domain.Event
	err := r.db.WithContext(ctx).
		Where("calendar_id = ? AND status = ?", calendarID, status).
		Order("start_time ASC").
		Find(&events).Error
	return events, err
}

func (r *calendarRepository) GetEventsByType(ctx context.Context, calendarID uint, eventType domain.EventType) ([]domain.Event, error) {
	var events []domain.Event
	err := r.db.WithContext(ctx).
		Where("calendar_id = ? AND type = ?", calendarID, eventType).
		Order("start_time ASC").
		Find(&events).Error
	return events, err
}
