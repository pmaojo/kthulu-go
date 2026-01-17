// @kthulu:store:participant
package store

import (
	"gorm.io/gorm"
	"tournament-api/internal/modules/participant/core"
)

type ParticipantRepository struct {
	db *gorm.DB
}

func NewParticipantRepository(db *gorm.DB) core.ParticipantRepository {
	return &ParticipantRepository{db: db}
}

func (r *ParticipantRepository) Create(entity *core.Participant) error {
	return r.db.Create(entity).Error
}

func (r *ParticipantRepository) GetByID(id uint) (*core.Participant, error) {
	var entity core.Participant
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *ParticipantRepository) Update(entity *core.Participant) error {
	return r.db.Save(entity).Error
}

func (r *ParticipantRepository) Delete(id uint) error {
	return r.db.Delete(&core.Participant{}, id).Error
}

func (r *ParticipantRepository) List(filter core.SearchFilter) ([]*core.Participant, error) {
	var entities []*core.Participant
	query := r.db.Model(&core.Participant{})

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
