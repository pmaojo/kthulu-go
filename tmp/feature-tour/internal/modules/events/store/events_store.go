// @kthulu:store:events
package store

import (
	"gorm.io/gorm"
	"feature-tour/internal/modules/events/core"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) core.EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(entity *core.Event) error {
	return r.db.Create(entity).Error
}

func (r *EventRepository) GetByID(id uint) (*core.Event, error) {
	var entity core.Event
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *EventRepository) Update(entity *core.Event) error {
	return r.db.Save(entity).Error
}

func (r *EventRepository) Delete(id uint) error {
	return r.db.Delete(&core.Event{}, id).Error
}

func (r *EventRepository) List(filter core.SearchFilter) ([]*core.Event, error) {
	var entities []*core.Event
	query := r.db.Model(&core.Event{})

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
