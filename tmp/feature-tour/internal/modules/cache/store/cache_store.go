// @kthulu:store:cache
package store

import (
	"gorm.io/gorm"
	"feature-tour/internal/modules/cache/core"
)

type CacheRepository struct {
	db *gorm.DB
}

func NewCacheRepository(db *gorm.DB) core.CacheRepository {
	return &CacheRepository{db: db}
}

func (r *CacheRepository) Create(entity *core.Cache) error {
	return r.db.Create(entity).Error
}

func (r *CacheRepository) GetByID(id uint) (*core.Cache, error) {
	var entity core.Cache
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *CacheRepository) Update(entity *core.Cache) error {
	return r.db.Save(entity).Error
}

func (r *CacheRepository) Delete(id uint) error {
	return r.db.Delete(&core.Cache{}, id).Error
}

func (r *CacheRepository) List(filter core.SearchFilter) ([]*core.Cache, error) {
	var entities []*core.Cache
	query := r.db.Model(&core.Cache{})

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
