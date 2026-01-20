// @kthulu:core:task
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

// Task represents a task entity
type Task struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
}

// TableName overrides the table name used by Task to `Tasks`
func (Task) TableName() string {
	return "Tasks"
}

// TaskRepository defines the repository interface
type TaskRepository interface {
	Create(entity *Task) error
	GetByID(id uint) (*Task, error)
	Update(entity *Task) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Task, error)
}

// TaskService defines the service interface  
type TaskService interface {
	CreateTask(entity *Task) error
	GetTaskByID(id uint) (*Task, error)
	UpdateTask(entity *Task) error
	DeleteTask(id uint) error
	ListTasks(filter SearchFilter) ([]*Task, error)
}
