// @kthulu:module:calendar
package domain

import (
	"time"
)

// CalendarType represents the type of calendar
type CalendarType string

const (
	CalendarTypePersonal CalendarType = "personal"
	CalendarTypeShared   CalendarType = "shared"
	CalendarTypeResource CalendarType = "resource"
	CalendarTypePublic   CalendarType = "public"
)

// Calendar represents a calendar container
type Calendar struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"not null;size:255"`
	Description string       `json:"description" gorm:"size:500"`
	Type        CalendarType `json:"type" gorm:"not null;size:20"`
	Color       string       `json:"color" gorm:"size:7"` // Hex color code
	IsActive    bool         `json:"isActive" gorm:"default:true"`
	OwnerID     uint         `json:"ownerId" gorm:"not null;index"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`

	// Relationships
	Owner  User    `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Events []Event `json:"events,omitempty" gorm:"foreignKey:CalendarID"`
}

// EventStatus represents the status of an event
type EventStatus string

const (
	EventStatusTentative EventStatus = "tentative"
	EventStatusConfirmed EventStatus = "confirmed"
	EventStatusCancelled EventStatus = "canceled"
)

// EventType represents the type of event
type EventType string

const (
	EventTypeAppointment EventType = "appointment"
	EventTypeMeeting     EventType = "meeting"
	EventTypeTask        EventType = "task"
	EventTypeReminder    EventType = "reminder"
	EventTypeBooking     EventType = "booking"
)

// Event represents a calendar event
type Event struct {
	ID             uint        `json:"id" gorm:"primaryKey"`
	CalendarID     uint        `json:"calendarId" gorm:"not null;index"`
	Title          string      `json:"title" gorm:"not null;size:255"`
	Description    string      `json:"description" gorm:"size:1000"`
	Location       string      `json:"location" gorm:"size:255"`
	StartTime      time.Time   `json:"startTime" gorm:"not null;index"`
	EndTime        time.Time   `json:"endTime" gorm:"not null;index"`
	AllDay         bool        `json:"allDay" gorm:"default:false"`
	Status         EventStatus `json:"status" gorm:"not null;size:20;default:'confirmed'"`
	Type           EventType   `json:"type" gorm:"not null;size:20;default:'appointment'"`
	IsRecurring    bool        `json:"isRecurring" gorm:"default:false"`
	RecurrenceRule string      `json:"recurrenceRule" gorm:"size:500"` // RRULE format
	CreatedByID    uint        `json:"createdById" gorm:"not null;index"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`

	// Relationships
	Calendar  Calendar   `json:"calendar,omitempty" gorm:"foreignKey:CalendarID"`
	CreatedBy User       `json:"createdBy,omitempty" gorm:"foreignKey:CreatedByID"`
	Attendees []Attendee `json:"attendees,omitempty" gorm:"foreignKey:EventID"`
	Reminders []Reminder `json:"reminders,omitempty" gorm:"foreignKey:EventID"`
}

// AttendeeStatus represents the status of an attendee
type AttendeeStatus string

const (
	AttendeeStatusPending   AttendeeStatus = "pending"
	AttendeeStatusAccepted  AttendeeStatus = "accepted"
	AttendeeStatusDeclined  AttendeeStatus = "declined"
	AttendeeStatusTentative AttendeeStatus = "tentative"
)

// Attendee represents an event attendee
type Attendee struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	EventID     uint           `json:"eventId" gorm:"not null;index"`
	UserID      *uint          `json:"userId" gorm:"index"`
	Email       string         `json:"email" gorm:"size:255"`
	Name        string         `json:"name" gorm:"size:255"`
	Status      AttendeeStatus `json:"status" gorm:"not null;size:20;default:'pending'"`
	IsOrganizer bool           `json:"isOrganizer" gorm:"default:false"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`

	// Relationships
	Event Event `json:"event,omitempty" gorm:"foreignKey:EventID"`
	User  *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// ReminderType represents the type of reminder
type ReminderType string

const (
	ReminderTypeEmail ReminderType = "email"
	ReminderTypeSMS   ReminderType = "sms"
	ReminderTypePush  ReminderType = "push"
)

// Reminder represents an event reminder
type Reminder struct {
	ID            uint         `json:"id" gorm:"primaryKey"`
	EventID       uint         `json:"eventId" gorm:"not null;index"`
	Type          ReminderType `json:"type" gorm:"not null;size:20"`
	MinutesBefore int          `json:"minutesBefore" gorm:"not null"`
	IsSent        bool         `json:"isSent" gorm:"default:false"`
	SentAt        *time.Time   `json:"sentAt"`
	CreatedAt     time.Time    `json:"createdAt"`

	// Relationships
	Event Event `json:"event,omitempty" gorm:"foreignKey:EventID"`
}

// AvailabilitySlot represents available time slots for booking
type AvailabilitySlot struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CalendarID  uint      `json:"calendarId" gorm:"not null;index"`
	StartTime   time.Time `json:"startTime" gorm:"not null;index"`
	EndTime     time.Time `json:"endTime" gorm:"not null;index"`
	IsAvailable bool      `json:"isAvailable" gorm:"default:true"`
	SlotType    string    `json:"slotType" gorm:"size:50"` // appointment, meeting, etc.
	Duration    int       `json:"duration"`                // Duration in minutes
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Relationships
	Calendar Calendar `json:"calendar,omitempty" gorm:"foreignKey:CalendarID"`
}

// Booking represents a booking made for an availability slot
type Booking struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	AvailabilitySlotID uint      `json:"availabilitySlotId" gorm:"not null;index"`
	EventID            *uint     `json:"eventId" gorm:"index"`
	BookedByID         uint      `json:"bookedById" gorm:"not null;index"`
	BookedForEmail     string    `json:"bookedForEmail" gorm:"size:255"`
	BookedForName      string    `json:"bookedForName" gorm:"size:255"`
	BookedForPhone     string    `json:"bookedForPhone" gorm:"size:50"`
	Notes              string    `json:"notes" gorm:"size:500"`
	Status             string    `json:"status" gorm:"not null;size:20;default:'confirmed'"` // confirmed, canceled, completed
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`

	// Relationships
	AvailabilitySlot AvailabilitySlot `json:"availabilitySlot,omitempty" gorm:"foreignKey:AvailabilitySlotID"`
	Event            *Event           `json:"event,omitempty" gorm:"foreignKey:EventID"`
	BookedBy         User             `json:"bookedBy,omitempty" gorm:"foreignKey:BookedByID"`
}

// IsOverlapping checks if this event overlaps with another time range
func (e *Event) IsOverlapping(startTime, endTime time.Time) bool {
	return e.StartTime.Before(endTime) && e.EndTime.After(startTime)
}

// Duration returns the duration of the event
func (e *Event) Duration() time.Duration {
	return e.EndTime.Sub(e.StartTime)
}

// IsInPast checks if the event is in the past
func (e *Event) IsInPast() bool {
	return e.EndTime.Before(time.Now())
}

// IsUpcoming checks if the event is upcoming (starts in the future)
func (e *Event) IsUpcoming() bool {
	return e.StartTime.After(time.Now())
}

// IsActive checks if the event is currently active
func (e *Event) IsActive() bool {
	now := time.Now()
	return e.StartTime.Before(now) && e.EndTime.After(now)
}

// IsBookable checks if the availability slot is available for booking
func (as *AvailabilitySlot) IsBookable() bool {
	return as.IsAvailable && as.StartTime.After(time.Now())
}

// GetDuration returns the duration of the availability slot
func (as *AvailabilitySlot) GetDuration() time.Duration {
	return as.EndTime.Sub(as.StartTime)
}
