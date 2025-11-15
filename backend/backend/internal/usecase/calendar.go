// @kthulu:module:calendar
package usecase

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// PaginatedResponse represents a paginated response
type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalPages int `json:"totalPages"`
}

// CalendarUseCase handles calendar management business logic
type CalendarUseCase struct {
	calendarRepo       repository.CalendarRepository
	userRepo           repository.UserRepository
	appointmentService *AppointmentService
	conflictChecker    *BookingConflictChecker
	logger             *zap.Logger
}

// NewCalendarUseCase creates a new calendar use case
func NewCalendarUseCase(
	calendarRepo repository.CalendarRepository,
	userRepo repository.UserRepository,
	logger *zap.Logger,
) *CalendarUseCase {
	// Default timezone - can be made configurable
	timezone := time.UTC

	return &CalendarUseCase{
		calendarRepo:       calendarRepo,
		userRepo:           userRepo,
		appointmentService: NewAppointmentService(timezone),
		conflictChecker:    NewBookingConflictChecker(15 * time.Minute), // 15 min buffer
		logger:             logger,
	}
}

// CreateCalendar creates a new calendar
func (uc *CalendarUseCase) CreateCalendar(ctx context.Context, req CreateCalendarRequest) (*domain.Calendar, error) {
	calendar := &domain.Calendar{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Color:       req.Color,
		OwnerID:     req.OwnerID,
		IsActive:    true,
	}

	if err := uc.calendarRepo.CreateCalendar(ctx, calendar); err != nil {
		uc.logger.Error("Failed to create calendar", zap.Error(err))
		return nil, err
	}

	uc.logger.Info("Calendar created", zap.String("name", calendar.Name), zap.Uint("ownerId", calendar.OwnerID))
	return calendar, nil
}

// GetCalendar retrieves a calendar by ID
func (uc *CalendarUseCase) GetCalendar(ctx context.Context, id uint) (*domain.Calendar, error) {
	calendar, err := uc.calendarRepo.GetCalendar(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get calendar", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return calendar, nil
}

// ListCalendars retrieves calendars for a user
func (uc *CalendarUseCase) ListCalendars(ctx context.Context, req ListCalendarsRequest) (*PaginatedResponse[domain.Calendar], error) {
	calendars, total, err := uc.calendarRepo.ListCalendars(ctx, req.OwnerID, req.Page, req.PageSize)
	if err != nil {
		uc.logger.Error("Failed to list calendars", zap.Error(err))
		return nil, err
	}

	return &PaginatedResponse[domain.Calendar]{
		Data:       calendars,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (total + req.PageSize - 1) / req.PageSize,
	}, nil
}

// CreateEvent creates a new event
func (uc *CalendarUseCase) CreateEvent(ctx context.Context, req CreateEventRequest) (*domain.Event, error) {
	// Validate calendar exists and user has access
	_, err := uc.calendarRepo.GetCalendar(ctx, req.CalendarID)
	if err != nil {
		return nil, fmt.Errorf("calendar not found: %w", err)
	}

	// Check for conflicts if requested
	if req.CheckConflicts {
		conflicts, err := uc.calendarRepo.GetEventsByTimeRange(ctx, req.CalendarID, req.StartTime, req.EndTime)
		if err != nil {
			return nil, fmt.Errorf("failed to check conflicts: %w", err)
		}
		if len(conflicts) > 0 {
			return nil, fmt.Errorf("event conflicts with existing events")
		}
	}

	event := &domain.Event{
		CalendarID:     req.CalendarID,
		Title:          req.Title,
		Description:    req.Description,
		Location:       req.Location,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		AllDay:         req.AllDay,
		Status:         domain.EventStatusConfirmed,
		Type:           req.Type,
		IsRecurring:    req.IsRecurring,
		RecurrenceRule: req.RecurrenceRule,
		CreatedByID:    req.CreatedByID,
	}

	if err := uc.calendarRepo.CreateEvent(ctx, event); err != nil {
		uc.logger.Error("Failed to create event", zap.Error(err))
		return nil, err
	}

	// Create attendees if provided
	if len(req.Attendees) > 0 {
		for _, attendeeReq := range req.Attendees {
			attendee := &domain.Attendee{
				EventID:     event.ID,
				UserID:      attendeeReq.UserID,
				Email:       attendeeReq.Email,
				Name:        attendeeReq.Name,
				Status:      domain.AttendeeStatusPending,
				IsOrganizer: attendeeReq.IsOrganizer,
			}
			if err := uc.calendarRepo.CreateAttendee(ctx, attendee); err != nil {
				uc.logger.Error("Failed to create attendee", zap.Error(err))
				// Don't fail the whole operation
			}
		}
	}

	// Create reminders if provided
	if len(req.Reminders) > 0 {
		for _, reminderReq := range req.Reminders {
			reminder := &domain.Reminder{
				EventID:       event.ID,
				Type:          reminderReq.Type,
				MinutesBefore: reminderReq.MinutesBefore,
			}
			if err := uc.calendarRepo.CreateReminder(ctx, reminder); err != nil {
				uc.logger.Error("Failed to create reminder", zap.Error(err))
				// Don't fail the whole operation
			}
		}
	}

	uc.logger.Info("Event created", zap.String("title", event.Title), zap.Uint("calendarId", event.CalendarID))
	return event, nil
}

// GetEvent retrieves an event by ID
func (uc *CalendarUseCase) GetEvent(ctx context.Context, id uint) (*domain.Event, error) {
	event, err := uc.calendarRepo.GetEvent(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get event", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return event, nil
}

// ListEvents retrieves events for a calendar within a time range
func (uc *CalendarUseCase) ListEvents(ctx context.Context, req ListEventsRequest) (*PaginatedResponse[domain.Event], error) {
	var events []domain.Event
	var total int
	var err error

	if req.StartTime != nil && req.EndTime != nil {
		events, err = uc.calendarRepo.GetEventsByTimeRange(ctx, req.CalendarID, *req.StartTime, *req.EndTime)
		total = len(events)
	} else {
		events, total, err = uc.calendarRepo.ListEvents(ctx, req.CalendarID, req.Page, req.PageSize)
	}

	if err != nil {
		uc.logger.Error("Failed to list events", zap.Error(err))
		return nil, err
	}

	return &PaginatedResponse[domain.Event]{
		Data:       events,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (total + req.PageSize - 1) / req.PageSize,
	}, nil
}

// GetAvailableSlots generates available time slots for a calendar
func (uc *CalendarUseCase) GetAvailableSlots(ctx context.Context, req GetAvailableSlotsRequest) ([]domain.AvailabilitySlot, error) {
	// Get existing events for the date
	startOfDay := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, req.Date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	existingEvents, err := uc.calendarRepo.GetEventsByTimeRange(ctx, req.CalendarID, startOfDay, endOfDay)
	if err != nil {
		uc.logger.Error("Failed to get existing events", zap.Error(err))
		return nil, err
	}

	// Generate available slots using appointment service
	slots := uc.appointmentService.GenerateAvailableSlots(req.Date, req.SlotDuration, existingEvents)

	// Set calendar ID for all slots
	for i := range slots {
		slots[i].CalendarID = req.CalendarID
	}

	return slots, nil
}

// BookSlot creates a booking for an available slot
func (uc *CalendarUseCase) BookSlot(ctx context.Context, req BookSlotRequest) (*domain.Booking, error) {
	// Create availability slot first
	slot := &domain.AvailabilitySlot{
		CalendarID:  req.CalendarID,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		IsAvailable: false, // Mark as booked
		SlotType:    req.SlotType,
		Duration:    int(req.EndTime.Sub(req.StartTime).Minutes()),
	}

	if err := uc.calendarRepo.CreateAvailabilitySlot(ctx, slot); err != nil {
		return nil, err
	}

	// Create the booking
	booking := &domain.Booking{
		AvailabilitySlotID: slot.ID,
		BookedByID:         req.BookedByID,
		BookedForEmail:     req.BookedForEmail,
		BookedForName:      req.BookedForName,
		BookedForPhone:     req.BookedForPhone,
		Notes:              req.Notes,
		Status:             "confirmed",
	}

	if err := uc.calendarRepo.CreateBooking(ctx, booking); err != nil {
		return nil, err
	}

	// Optionally create an event for the booking
	if req.CreateEvent {
		event := &domain.Event{
			CalendarID:  req.CalendarID,
			Title:       fmt.Sprintf("Appointment with %s", req.BookedForName),
			Description: req.Notes,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			Status:      domain.EventStatusConfirmed,
			Type:        domain.EventTypeBooking,
			CreatedByID: req.BookedByID,
		}

		if err := uc.calendarRepo.CreateEvent(ctx, event); err != nil {
			uc.logger.Error("Failed to create event for booking", zap.Error(err))
			// Don't fail the booking
		} else {
			booking.EventID = &event.ID
			uc.calendarRepo.UpdateBooking(ctx, booking)
		}
	}

	uc.logger.Info("Slot booked",
		zap.Uint("calendarId", req.CalendarID),
		zap.String("bookedFor", req.BookedForName))

	return booking, nil
}

// IsBusinessDay checks if a date is a business day
func (uc *CalendarUseCase) IsBusinessDay(ctx context.Context, date time.Time) bool {
	return uc.appointmentService.IsWorkingDay(date)
}

// ValidateAppointmentTime validates if an appointment time is valid
func (uc *CalendarUseCase) ValidateAppointmentTime(ctx context.Context, startTime, endTime time.Time) error {
	return uc.appointmentService.ValidateAppointmentTime(startTime, endTime)
}

// Request/Response DTOs
type CreateCalendarRequest struct {
	Name        string              `json:"name" validate:"required,max=255"`
	Description string              `json:"description" validate:"max=500"`
	Type        domain.CalendarType `json:"type" validate:"required"`
	Color       string              `json:"color" validate:"max=7"`
	OwnerID     uint                `json:"ownerId" validate:"required"`
}

type ListCalendarsRequest struct {
	OwnerID  uint `json:"ownerId" validate:"required"`
	Page     int  `json:"page" validate:"min=1"`
	PageSize int  `json:"pageSize" validate:"min=1,max=100"`
}

type CreateEventRequest struct {
	CalendarID     uint                    `json:"calendarId" validate:"required"`
	Title          string                  `json:"title" validate:"required,max=255"`
	Description    string                  `json:"description" validate:"max=1000"`
	Location       string                  `json:"location" validate:"max=255"`
	StartTime      time.Time               `json:"startTime" validate:"required"`
	EndTime        time.Time               `json:"endTime" validate:"required"`
	AllDay         bool                    `json:"allDay"`
	Type           domain.EventType        `json:"type" validate:"required"`
	IsRecurring    bool                    `json:"isRecurring"`
	RecurrenceRule string                  `json:"recurrenceRule" validate:"max=500"`
	CreatedByID    uint                    `json:"createdById" validate:"required"`
	CheckConflicts bool                    `json:"checkConflicts"`
	Attendees      []CreateAttendeeRequest `json:"attendees"`
	Reminders      []CreateReminderRequest `json:"reminders"`
}

type CreateAttendeeRequest struct {
	UserID      *uint  `json:"userId"`
	Email       string `json:"email" validate:"required,email"`
	Name        string `json:"name" validate:"required,max=255"`
	IsOrganizer bool   `json:"isOrganizer"`
}

type CreateReminderRequest struct {
	Type          domain.ReminderType `json:"type" validate:"required"`
	MinutesBefore int                 `json:"minutesBefore" validate:"required,min=0"`
}

type ListEventsRequest struct {
	CalendarID uint       `json:"calendarId" validate:"required"`
	StartTime  *time.Time `json:"startTime"`
	EndTime    *time.Time `json:"endTime"`
	Page       int        `json:"page" validate:"min=1"`
	PageSize   int        `json:"pageSize" validate:"min=1,max=100"`
}

type GetAvailableSlotsRequest struct {
	CalendarID   uint          `json:"calendarId" validate:"required"`
	Date         time.Time     `json:"date" validate:"required"`
	SlotDuration time.Duration `json:"slotDuration" validate:"required"`
}

type BookSlotRequest struct {
	CalendarID     uint      `json:"calendarId" validate:"required"`
	StartTime      time.Time `json:"startTime" validate:"required"`
	EndTime        time.Time `json:"endTime" validate:"required"`
	SlotType       string    `json:"slotType" validate:"required"`
	BookedByID     uint      `json:"bookedById" validate:"required"`
	BookedForEmail string    `json:"bookedForEmail" validate:"required,email"`
	BookedForName  string    `json:"bookedForName" validate:"required,max=255"`
	BookedForPhone string    `json:"bookedForPhone" validate:"max=50"`
	Notes          string    `json:"notes" validate:"max=500"`
	CreateEvent    bool      `json:"createEvent"`
}
