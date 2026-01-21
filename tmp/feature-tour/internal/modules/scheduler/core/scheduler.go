// @kthulu:core:scheduler
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

// Scheduler represents a scheduler entity
type Scheduler struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Name string `json:"name" gorm:"name"`
}

// TableName overrides the table name used by Scheduler to `Schedulers`
func (Scheduler) TableName() string {
	return "Schedulers"
}

// SchedulerRepository defines the repository interface
type SchedulerRepository interface {
	Create(entity *Scheduler) error
	GetByID(id uint) (*Scheduler, error)
	Update(entity *Scheduler) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Scheduler, error)
}

// SchedulerService defines the service interface  
type SchedulerService interface {
	CreateScheduler(entity *Scheduler) error
	GetSchedulerByID(id uint) (*Scheduler, error)
	UpdateScheduler(entity *Scheduler) error
	DeleteScheduler(id uint) error
	ListSchedulers(filter SearchFilter) ([]*Scheduler, error)
}
