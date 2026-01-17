// @kthulu:store:matches
package store

import (
	"gorm.io/gorm"
	"tournament-app/internal/modules/matches/core"
)

type MatchRepository struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) core.MatchRepository {
	return &MatchRepository{db: db}
}

func (r *MatchRepository) Create(entity *core.Match) error {
	return r.db.Create(entity).Error
}

func (r *MatchRepository) GetByID(id uint) (*core.Match, error) {
	var entity core.Match
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *MatchRepository) Update(entity *core.Match) error {
	return r.db.Save(entity).Error
}

func (r *MatchRepository) Delete(id uint) error {
	return r.db.Delete(&core.Match{}, id).Error
}

func (r *MatchRepository) List(filter core.SearchFilter) ([]*core.Match, error) {
	var entities []*core.Match
	query := r.db.Model(&core.Match{})

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
