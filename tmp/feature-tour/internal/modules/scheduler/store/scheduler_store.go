// @kthulu:store:scheduler
package store

import (
	"gorm.io/gorm"
	"feature-tour/internal/modules/scheduler/core"
)

type SchedulerRepository struct {
	db *gorm.DB
}

func NewSchedulerRepository(db *gorm.DB) core.SchedulerRepository {
	return &SchedulerRepository{db: db}
}

func (r *SchedulerRepository) Create(entity *core.Scheduler) error {
	return r.db.Create(entity).Error
}

func (r *SchedulerRepository) GetByID(id uint) (*core.Scheduler, error) {
	var entity core.Scheduler
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *SchedulerRepository) Update(entity *core.Scheduler) error {
	return r.db.Save(entity).Error
}

func (r *SchedulerRepository) Delete(id uint) error {
	return r.db.Delete(&core.Scheduler{}, id).Error
}

func (r *SchedulerRepository) List(filter core.SearchFilter) ([]*core.Scheduler, error) {
	var entities []*core.Scheduler
	query := r.db.Model(&core.Scheduler{})

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
