// @kthulu:core:tournaments
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

// Tournament represents a tournaments entity
type Tournament struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Name string `json:"name" gorm:"name"`
	Status string `json:"status" gorm:"status"`
	Type string `json:"type" gorm:"type"`
}

// TableName overrides the table name used by Tournament to `Tournaments`
func (Tournament) TableName() string {
	return "Tournaments"
}

// TournamentRepository defines the repository interface
type TournamentRepository interface {
	Create(entity *Tournament) error
	GetByID(id uint) (*Tournament, error)
	Update(entity *Tournament) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Tournament, error)
}

// TournamentService defines the service interface  
type TournamentService interface {
	CreateTournament(entity *Tournament) error
	GetTournamentByID(id uint) (*Tournament, error)
	UpdateTournament(entity *Tournament) error
	DeleteTournament(id uint) error
	ListTournaments(filter SearchFilter) ([]*Tournament, error)
}
