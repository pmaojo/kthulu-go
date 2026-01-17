package store

import (
	"gorm.io/gorm"
	"tournament-api/internal/modules/tournamentv2/core"
)

type tournamentV2Store struct {
	db *gorm.DB
}

func NewTournamentV2Store(db *gorm.DB) core.TournamentV2Repository {
	return &tournamentV2Store{db: db}
}

func (s *tournamentV2Store) Create(entity *core.TournamentV2) error {
	return s.db.Create(entity).Error
}

func (s *tournamentV2Store) GetByID(id uint) (*core.TournamentV2, error) {
	var entity core.TournamentV2
	if err := s.db.First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (s *tournamentV2Store) Update(entity *core.TournamentV2) error {
	return s.db.Save(entity).Error
}

func (s *tournamentV2Store) Delete(id uint) error {
	return s.db.Delete(&core.TournamentV2{}, id).Error
}

func (s *tournamentV2Store) List(filter core.SearchFilter) ([]*core.TournamentV2, error) {
	var entities []*core.TournamentV2
	db := s.db
	if filter.Query != "" {
		db = db.Where("name LIKE ?", "%"+filter.Query+"%")
	}
	if filter.Limit > 0 {
		db = db.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		db = db.Offset(filter.Offset)
	}
	if err := db.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}
