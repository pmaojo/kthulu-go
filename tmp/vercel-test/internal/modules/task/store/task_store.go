// @kthulu:store:task
package store

import (
	"gorm.io/gorm"
	"vercel-test/internal/modules/task/core"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) core.TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(entity *core.Task) error {
	return r.db.Create(entity).Error
}

func (r *TaskRepository) GetByID(id uint) (*core.Task, error) {
	var entity core.Task
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *TaskRepository) Update(entity *core.Task) error {
	return r.db.Save(entity).Error
}

func (r *TaskRepository) Delete(id uint) error {
	return r.db.Delete(&core.Task{}, id).Error
}

func (r *TaskRepository) List(filter core.SearchFilter) ([]*core.Task, error) {
	var entities []*core.Task
	query := r.db.Model(&core.Task{})

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
