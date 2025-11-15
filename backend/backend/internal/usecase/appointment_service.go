// @kthulu:module:calendar
package usecase

import (
	"fmt"
	"time"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
)

// AppointmentService provides appointment scheduling functionality
// Inspired by github.com/juanjoss/appointments-system
type AppointmentService struct {
	workingHours domain.WorkingHours
	timezone     *time.Location
}

// NewAppointmentService creates a new appointment service
func NewAppointmentService(timezone *time.Location) *AppointmentService {
	if timezone == nil {
		timezone = time.UTC
	}

	return &AppointmentService{
		workingHours: domain.DefaultWorkingHours(),
		timezone:     timezone,
	}
}

// SetWorkingHours updates the working hours
func (as *AppointmentService) SetWorkingHours(workingHours domain.WorkingHours) {
	as.workingHours = workingHours
}

// IsWorkingDay checks if a given date is a working day
func (as *AppointmentService) IsWorkingDay(date time.Time) bool {
	return as.workingHours.IsWorkingDay(date)
}

// GetWorkingHours returns the working hours for a specific date
func (as *AppointmentService) GetWorkingHours(date time.Time) (start, end time.Time, hasBreak bool, breakStart, breakEnd time.Time) {
	return as.workingHours.GetWorkingHoursForDate(date, as.timezone)
}

// GenerateAvailableSlots generates available appointment slots for a given date
func (as *AppointmentService) GenerateAvailableSlots(date time.Time, slotDuration time.Duration, existingAppointments []domain.Event) []domain.AvailabilitySlot {
	if !as.IsWorkingDay(date) {
		return []domain.AvailabilitySlot{}
	}

	workStart, workEnd, hasBreak, breakStart, breakEnd := as.GetWorkingHours(date)
	var slots []domain.AvailabilitySlot

	// Generate morning slots (before break if exists)
	endTime := workEnd
	if hasBreak {
		endTime = breakStart
	}

	slots = append(slots, as.generateSlotsForPeriod(workStart, endTime, slotDuration, existingAppointments)...)

	// Generate afternoon slots (after break if exists)
	if hasBreak {
		slots = append(slots, as.generateSlotsForPeriod(breakEnd, workEnd, slotDuration, existingAppointments)...)
	}

	return slots
}

// generateSlotsForPeriod generates slots for a specific time period
func (as *AppointmentService) generateSlotsForPeriod(start, end time.Time, slotDuration time.Duration, existingAppointments []domain.Event) []domain.AvailabilitySlot {
	var slots []domain.AvailabilitySlot
	current := start

	for current.Add(slotDuration).Before(end) || current.Add(slotDuration).Equal(end) {
		slotEnd := current.Add(slotDuration)

		// Check if this slot conflicts with existing appointments
		isAvailable := true
		for _, appointment := range existingAppointments {
			if appointment.IsOverlapping(current, slotEnd) {
				isAvailable = false
				break
			}
		}

		if isAvailable {
			slots = append(slots, domain.AvailabilitySlot{
				StartTime:   current,
				EndTime:     slotEnd,
				IsAvailable: true,
				SlotType:    "appointment",
				Duration:    int(slotDuration.Minutes()),
			})
		}

		current = current.Add(slotDuration)
	}

	return slots
}

// ValidateAppointmentTime checks if an appointment time is valid
func (as *AppointmentService) ValidateAppointmentTime(startTime, endTime time.Time) error {
	return as.workingHours.ValidateAppointmentTime(startTime, endTime, as.timezone)
}

// GetNextAvailableSlot finds the next available slot after a given time
func (as *AppointmentService) GetNextAvailableSlot(after time.Time, duration time.Duration, existingAppointments []domain.Event) (*domain.AvailabilitySlot, error) {
	// Start from the next day if the time is in the past or today after working hours
	current := after
	if current.Before(time.Now()) {
		current = time.Now()
	}

	// Look for available slots in the next 30 days
	for i := 0; i < 30; i++ {
		checkDate := current.AddDate(0, 0, i)
		slots := as.GenerateAvailableSlots(checkDate, duration, existingAppointments)

		for _, slot := range slots {
			if slot.StartTime.After(current) {
				return &slot, nil
			}
		}
	}

	return nil, fmt.Errorf("no available slots found in the next 30 days")
}

// BookingConflictChecker checks for booking conflicts
type BookingConflictChecker struct {
	bufferTime time.Duration // Buffer time between appointments
}

// NewBookingConflictChecker creates a new conflict checker
func NewBookingConflictChecker(bufferTime time.Duration) *BookingConflictChecker {
	return &BookingConflictChecker{
		bufferTime: bufferTime,
	}
}

// CheckConflicts checks if a new booking conflicts with existing ones
func (bcc *BookingConflictChecker) CheckConflicts(newStart, newEnd time.Time, existingBookings []domain.Event) []domain.Event {
	var conflicts []domain.Event

	// Add buffer time to the new booking
	bufferedStart := newStart.Add(-bcc.bufferTime)
	bufferedEnd := newEnd.Add(bcc.bufferTime)

	for _, booking := range existingBookings {
		if booking.IsOverlapping(bufferedStart, bufferedEnd) {
			conflicts = append(conflicts, booking)
		}
	}

	return conflicts
}

// HasConflicts checks if there are any conflicts
func (bcc *BookingConflictChecker) HasConflicts(newStart, newEnd time.Time, existingBookings []domain.Event) bool {
	conflicts := bcc.CheckConflicts(newStart, newEnd, existingBookings)
	return len(conflicts) > 0
}
