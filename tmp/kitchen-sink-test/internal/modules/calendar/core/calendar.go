// @kthulu:core:calendar
package core

import (
	"time"
	
)

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// Calendar represents a calendar entity
type Calendar struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Name string `json:"name" gorm:"name"`
	Start_time string `json:"start_time" gorm:"start_time"`
	End_time string `json:"end_time" gorm:"end_time"`
	Location string `json:"location" gorm:"location"`
}

// TableName overrides the table name used by Calendar to `Calendars`
func (Calendar) TableName() string {
	return "Calendars"
}

// CalendarRepository defines the repository interface
type CalendarRepository interface {
	Create(entity *Calendar) error
	GetByID(id uint) (*Calendar, error)
	Update(entity *Calendar) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Calendar, error)
}

// CalendarService defines the service interface  
type CalendarService interface {
	CreateCalendar(entity *Calendar) error
	GetCalendarByID(id uint) (*Calendar, error)
	UpdateCalendar(entity *Calendar) error
	DeleteCalendar(id uint) error
	ListCalendars(filter SearchFilter) ([]*Calendar, error)
}
