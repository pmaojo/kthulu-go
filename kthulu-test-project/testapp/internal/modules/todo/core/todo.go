// @kthulu:core:todo
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

// Todo represents a todo entity
type Todo struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Title string `json:"title" gorm:"title"`
	Completed bool `json:"completed" gorm:"completed"`
}

// TableName overrides the table name used by Todo to `Todos`
func (Todo) TableName() string {
	return "Todos"
}

// TodoRepository defines the repository interface
type TodoRepository interface {
	Create(entity *Todo) error
	GetByID(id uint) (*Todo, error)
	Update(entity *Todo) error
	Delete(id uint) error
	List(filter SearchFilter) ([]*Todo, error)
}

// TodoService defines the service interface  
type TodoService interface {
	CreateTodo(entity *Todo) error
	GetTodoByID(id uint) (*Todo, error)
	UpdateTodo(entity *Todo) error
	DeleteTodo(id uint) error
	ListTodos(filter SearchFilter) ([]*Todo, error)
}
