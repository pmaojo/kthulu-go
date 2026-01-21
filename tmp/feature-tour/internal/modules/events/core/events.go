// @kthulu:core:events
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

// Event represents a events entity
type Event struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Name string `json:"name" gorm:"name"`
}

// TableName overrides the table name used by Event to `Events`
func (Event) TableName() string {
	return "Events"
}

// EventRepository defines the repository interface
type EventRepository interface {
	Create(entity *Event) error
	GetByID(id uint) (*Event, error)
	Update(entity *Event) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Event, error)
}

// EventService defines the service interface  
type EventService interface {
	CreateEvent(entity *Event) error
	GetEventByID(id uint) (*Event, error)
	UpdateEvent(entity *Event) error
	DeleteEvent(id uint) error
	ListEvents(filter SearchFilter) ([]*Event, error)
}
