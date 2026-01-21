// @kthulu:store:invoice
package store

import (
	"gorm.io/gorm"
	"feature-tour/internal/modules/invoice/core"
)

type InvoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) core.InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Create(entity *core.Invoice) error {
	return r.db.Create(entity).Error
}

func (r *InvoiceRepository) GetByID(id uint) (*core.Invoice, error) {
	var entity core.Invoice
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *InvoiceRepository) Update(entity *core.Invoice) error {
	return r.db.Save(entity).Error
}

func (r *InvoiceRepository) Delete(id uint) error {
	return r.db.Delete(&core.Invoice{}, id).Error
}

func (r *InvoiceRepository) List(filter core.SearchFilter) ([]*core.Invoice, error) {
	var entities []*core.Invoice
	query := r.db.Model(&core.Invoice{})

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
