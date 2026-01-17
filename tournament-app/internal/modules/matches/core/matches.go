// @kthulu:core:matches
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

// Match represents a matches entity
type Match struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Tournament_id string `json:"tournament_id" gorm:"tournament_id"`
	Participant1_id string `json:"participant1_id" gorm:"participant1_id"`
	Participant2_id string `json:"participant2_id" gorm:"participant2_id"`
	Score1 int `json:"score1" gorm:"score1"`
	Score2 int `json:"score2" gorm:"score2"`
	Winner_id string `json:"winner_id" gorm:"winner_id"`
}

// TableName overrides the table name used by Match to `Matches`
func (Match) TableName() string {
	return "Matches"
}

// MatchRepository defines the repository interface
type MatchRepository interface {
	Create(entity *Match) error
	GetByID(id uint) (*Match, error)
	Update(entity *Match) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Match, error)
}

// MatchService defines the service interface  
type MatchService interface {
	CreateMatch(entity *Match) error
	GetMatchByID(id uint) (*Match, error)
	UpdateMatch(entity *Match) error
	DeleteMatch(id uint) error
	ListMatches(filter SearchFilter) ([]*Match, error)
}
