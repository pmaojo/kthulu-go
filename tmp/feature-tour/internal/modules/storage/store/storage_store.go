// @kthulu:store:storage
package store

import (
	"gorm.io/gorm"
	"feature-tour/internal/modules/storage/core"
)

type StorageRepository struct {
	db *gorm.DB
}

func NewStorageRepository(db *gorm.DB) core.StorageRepository {
	return &StorageRepository{db: db}
}

func (r *StorageRepository) Create(entity *core.Storage) error {
	return r.db.Create(entity).Error
}

func (r *StorageRepository) GetByID(id uint) (*core.Storage, error) {
	var entity core.Storage
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *StorageRepository) Update(entity *core.Storage) error {
	return r.db.Save(entity).Error
}

func (r *StorageRepository) Delete(id uint) error {
	return r.db.Delete(&core.Storage{}, id).Error
}

func (r *StorageRepository) List(filter core.SearchFilter) ([]*core.Storage, error) {
	var entities []*core.Storage
	query := r.db.Model(&core.Storage{})

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
