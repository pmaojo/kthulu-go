// @kthulu:store:inventory
package store

import (
	"gorm.io/gorm"
	"kitchen-sink-test/internal/modules/inventory/core"
)

type InventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) core.InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) Create(entity *core.Inventory) error {
	return r.db.Create(entity).Error
}

func (r *InventoryRepository) GetByID(id uint) (*core.Inventory, error) {
	var entity core.Inventory
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *InventoryRepository) Update(entity *core.Inventory) error {
	return r.db.Save(entity).Error
}

func (r *InventoryRepository) Delete(id uint) error {
	return r.db.Delete(&core.Inventory{}, id).Error
}

func (r *InventoryRepository) List(filter core.SearchFilter) ([]*core.Inventory, error) {
	var entities []*core.Inventory
	query := r.db.Model(&core.Inventory{})

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
