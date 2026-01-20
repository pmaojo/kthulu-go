// @kthulu:store:calendar
package store

import (
	"gorm.io/gorm"
	"kitchen-sink-test/internal/modules/calendar/core"
)

type CalendarRepository struct {
	db *gorm.DB
}

func NewCalendarRepository(db *gorm.DB) core.CalendarRepository {
	return &CalendarRepository{db: db}
}

func (r *CalendarRepository) Create(entity *core.Calendar) error {
	return r.db.Create(entity).Error
}

func (r *CalendarRepository) GetByID(id uint) (*core.Calendar, error) {
	var entity core.Calendar
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *CalendarRepository) Update(entity *core.Calendar) error {
	return r.db.Save(entity).Error
}

func (r *CalendarRepository) Delete(id uint) error {
	return r.db.Delete(&core.Calendar{}, id).Error
}

func (r *CalendarRepository) List(filter core.SearchFilter) ([]*core.Calendar, error) {
	var entities []*core.Calendar
	query := r.db.Model(&core.Calendar{})

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
