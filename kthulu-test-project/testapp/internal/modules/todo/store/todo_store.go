// @kthulu:store:todo
package store

import (
	"gorm.io/gorm"
	"testapp/internal/modules/todo/core"
)

type TodoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) core.TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) Create(entity *core.Todo) error {
	return r.db.Create(entity).Error
}

func (r *TodoRepository) GetByID(id uint) (*core.Todo, error) {
	var entity core.Todo
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *TodoRepository) Update(entity *core.Todo) error {
	return r.db.Save(entity).Error
}

func (r *TodoRepository) Delete(id uint) error {
	return r.db.Delete(&core.Todo{}, id).Error
}

func (r *TodoRepository) List(filter core.SearchFilter) ([]*core.Todo, error) {
	var entities []*core.Todo
	query := r.db.Model(&core.Todo{})

	if filter.Query != "" {
		// Basic search implementation
		// query = query.Where("name LIKE ?", "%"+filter.Query+"%")
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Find(&entities).Error
	return entities, err
}
