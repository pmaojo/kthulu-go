// @kthulu:core:tournamentv2
package core

import (
	"time"
)

// SearchFilter represents search criteria
type SearchFilter struct {
	Query  string
	Limit  int
	Offset int
}

// TournamentV2 represents an enhanced tournament entity
type TournamentV2 struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name" gorm:"name"`
	Description string    `json:"description" gorm:"description"`
	StartDate   time.Time `json:"start_date" gorm:"start_date"`
	EndDate     time.Time `json:"end_date" gorm:"end_date"`
	Status      string    `json:"status" gorm:"status"`
	MaxTeams    int       `json:"max_teams" gorm:"max_teams"`
	PrizePool   float64   `json:"prize_pool" gorm:"prize_pool"`
	Rules       string    `json:"rules" gorm:"rules"`
}

// TableName overrides the table name used by TournamentV2 to `tournaments_v2`
func (TournamentV2) TableName() string {
	return "tournaments_v2"
}

// TournamentV2Repository defines the repository interface
type TournamentV2Repository interface {
	Create(entity *TournamentV2) error
	GetByID(id uint) (*TournamentV2, error)
	Update(entity *TournamentV2) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*TournamentV2, error)
}

// TournamentV2Service defines the service interface
type TournamentV2Service interface {
	CreateTournament(entity *TournamentV2) error
	GetTournamentByID(id uint) (*TournamentV2, error)
	UpdateTournament(entity *TournamentV2) error
	DeleteTournament(id uint) error
	ListTournaments(filter SearchFilter) ([]*TournamentV2, error)
}
