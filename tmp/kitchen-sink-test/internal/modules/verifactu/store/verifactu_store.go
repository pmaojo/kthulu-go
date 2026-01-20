// @kthulu:store:verifactu
package store

import (
	"gorm.io/gorm"
	"kitchen-sink-test/internal/modules/verifactu/core"
)

type VerifactuRepository struct {
	db *gorm.DB
}

func NewVerifactuRepository(db *gorm.DB) core.VerifactuRepository {
	return &VerifactuRepository{db: db}
}

func (r *VerifactuRepository) Create(entity *core.Verifactu) error {
	return r.db.Create(entity).Error
}

func (r *VerifactuRepository) GetByID(id uint) (*core.Verifactu, error) {
	var entity core.Verifactu
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *VerifactuRepository) Update(entity *core.Verifactu) error {
	return r.db.Save(entity).Error
}

func (r *VerifactuRepository) Delete(id uint) error {
	return r.db.Delete(&core.Verifactu{}, id).Error
}

func (r *VerifactuRepository) List(filter core.SearchFilter) ([]*core.Verifactu, error) {
	var entities []*core.Verifactu
	query := r.db.Model(&core.Verifactu{})

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
