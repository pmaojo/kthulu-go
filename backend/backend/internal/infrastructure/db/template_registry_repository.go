// @kthulu:module:templates
package db

import (
	"context"
	"errors"

	"backend/internal/domain"
	"backend/internal/repository"

	"gorm.io/gorm"
)

// TemplateRegistryModel represents the database model for template registries
type TemplateRegistryModel struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null;uniqueIndex"`
	URL       string `gorm:"not null;unique"`
	CreatedAt string `gorm:"type:text"`
	UpdatedAt string `gorm:"type:text"`
}

// TableName specifies the table name for TemplateRegistryModel
func (TemplateRegistryModel) TableName() string {
	return "template_registries"
}

// ToDomain converts TemplateRegistryModel to domain.TemplateRegistry
func (t *TemplateRegistryModel) ToDomain() (*domain.TemplateRegistry, error) {
	registry := &domain.TemplateRegistry{
		ID:   t.ID,
		Name: t.Name,
		URL:  t.URL,
	}

	return registry, nil
}

// FromDomain converts domain.TemplateRegistry to TemplateRegistryModel
func (t *TemplateRegistryModel) FromDomain(registry *domain.TemplateRegistry) {
	t.ID = registry.ID
	t.Name = registry.Name
	t.URL = registry.URL
}

// TemplateRegistryRepository provides a database-backed implementation of repository.TemplateRegistryRepository.
type TemplateRegistryRepository struct {
	db *gorm.DB
}

// NewTemplateRegistryRepository creates a new instance bound to a Gorm database.
func NewTemplateRegistryRepository(db *gorm.DB) repository.TemplateRegistryRepository {
	return &TemplateRegistryRepository{db: db}
}

// Create persists a new template registry.
func (r *TemplateRegistryRepository) Create(ctx context.Context, registry *domain.TemplateRegistry) error {
	model := &TemplateRegistryModel{}
	model.FromDomain(registry)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	registry.ID = model.ID
	return nil
}

// FindByID retrieves a template registry by ID.
func (r *TemplateRegistryRepository) FindByID(ctx context.Context, id uint) (*domain.TemplateRegistry, error) {
	var model TemplateRegistryModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template registry not found")
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindByName retrieves a template registry by name.
func (r *TemplateRegistryRepository) FindByName(ctx context.Context, name string) (*domain.TemplateRegistry, error) {
	var model TemplateRegistryModel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template registry not found")
		}
		return nil, err
	}

	return model.ToDomain()
}

// Update saves template registry changes.
func (r *TemplateRegistryRepository) Update(ctx context.Context, registry *domain.TemplateRegistry) error {
	model := &TemplateRegistryModel{}
	model.FromDomain(registry)

	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a template registry by ID.
func (r *TemplateRegistryRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&TemplateRegistryModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("template registry not found")
	}
	return nil
}

// List retrieves template registries with pagination.
func (r *TemplateRegistryRepository) List(ctx context.Context, limit, offset int) ([]*domain.TemplateRegistry, error) {
	var models []TemplateRegistryModel
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}

	registries := make([]*domain.TemplateRegistry, len(models))
	for i, model := range models {
		registry, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		registries[i] = registry
	}

	return registries, nil
}

// Count returns the total number of template registries.
func (r *TemplateRegistryRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TemplateRegistryModel{}).Count(&count).Error
	return count, err
}

// ExistsByName checks if a template registry exists with the given name.
func (r *TemplateRegistryRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TemplateRegistryModel{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// ExistsByID checks if a template registry exists with the given ID.
func (r *TemplateRegistryRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TemplateRegistryModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Ensure TemplateRegistryRepository implements the interface
var _ repository.TemplateRegistryRepository = (*TemplateRegistryRepository)(nil)
