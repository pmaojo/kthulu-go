// @kthulu:store:debugmod
package store

import (
	"gorm.io/gorm"
	"tournament-app/internal/modules/debugmod/core"
)

type DebugmodRepository struct {
	db *gorm.DB
}

func NewDebugmodRepository(db *gorm.DB) core.DebugmodRepository {
	return &DebugmodRepository{db: db}
}

func (r *DebugmodRepository) Create(entity *core.Debugmod) error {
	return r.db.Create(entity).Error
}

func (r *DebugmodRepository) GetByID(id uint) (*core.Debugmod, error) {
	var entity core.Debugmod
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *DebugmodRepository) Update(entity *core.Debugmod) error {
	return r.db.Save(entity).Error
}

func (r *DebugmodRepository) Delete(id uint) error {
	return r.db.Delete(&core.Debugmod{}, id).Error
}

func (r *DebugmodRepository) List(filter core.SearchFilter) ([]*core.Debugmod, error) {
	var entities []*core.Debugmod
	query := r.db.Model(&core.Debugmod{})

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
