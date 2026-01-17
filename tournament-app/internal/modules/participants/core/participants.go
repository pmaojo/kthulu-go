// @kthulu:core:participants
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

// Participant represents a participants entity
type Participant struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Name string `json:"name" gorm:"name"`
	Team string `json:"team" gorm:"team"`
}

// TableName overrides the table name used by Participant to `Participants`
func (Participant) TableName() string {
	return "Participants"
}

// ParticipantRepository defines the repository interface
type ParticipantRepository interface {
	Create(entity *Participant) error
	GetByID(id uint) (*Participant, error)
	Update(entity *Participant) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Participant, error)
}

// ParticipantService defines the service interface  
type ParticipantService interface {
	CreateParticipant(entity *Participant) error
	GetParticipantByID(id uint) (*Participant, error)
	UpdateParticipant(entity *Participant) error
	DeleteParticipant(id uint) error
	ListParticipants(filter SearchFilter) ([]*Participant, error)
}
