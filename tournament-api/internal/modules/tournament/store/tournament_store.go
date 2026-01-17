// @kthulu:store:tournament
package store

import (
	"gorm.io/gorm"
	"tournament-api/internal/modules/tournament/core"
)

type TournamentRepository struct {
	db *gorm.DB
}

func NewTournamentRepository(db *gorm.DB) core.TournamentRepository {
	return &TournamentRepository{db: db}
}

func (r *TournamentRepository) Create(entity *core.Tournament) error {
	return r.db.Create(entity).Error
}

func (r *TournamentRepository) GetByID(id uint) (*core.Tournament, error) {
	var entity core.Tournament
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *TournamentRepository) Update(entity *core.Tournament) error {
	return r.db.Save(entity).Error
}

func (r *TournamentRepository) Delete(id uint) error {
	return r.db.Delete(&core.Tournament{}, id).Error
}

func (r *TournamentRepository) List(filter core.SearchFilter) ([]*core.Tournament, error) {
	var entities []*core.Tournament
	query := r.db.Model(&core.Tournament{})

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
