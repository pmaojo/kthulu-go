// @kthulu:module:access
package db

import (
	"context"
	"errors"
	"time"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"

	"gorm.io/gorm"
)

// PermissionModel represents the database model for permissions
// Provides structure for Gorm ORM mapping.
type PermissionModel struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	Resource    string `gorm:"not null"`
	Action      string `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName specifies the table name for PermissionModel
func (PermissionModel) TableName() string {
	return "permissions"
}

// ToDomain converts PermissionModel to domain.Permission
func (p *PermissionModel) ToDomain() (*domain.Permission, error) {
	return &domain.Permission{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Resource:    p.Resource,
		Action:      p.Action,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}, nil
}

// FromDomain converts domain.Permission to PermissionModel
func (p *PermissionModel) FromDomain(permission *domain.Permission) {
	p.ID = permission.ID
	p.Name = permission.Name
	p.Description = permission.Description
	p.Resource = permission.Resource
	p.Action = permission.Action
	p.CreatedAt = permission.CreatedAt
	p.UpdatedAt = permission.UpdatedAt
}

// PermissionRepository provides a database-backed implementation of repository.PermissionRepository.
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new instance bound to a Gorm database.
func NewPermissionRepository(db *gorm.DB) repository.PermissionRepository {
	return &PermissionRepository{db: db}
}

// Create persists a new permission.
func (r *PermissionRepository) Create(ctx context.Context, permission *domain.Permission) error {
	model := &PermissionModel{}
	model.FromDomain(permission)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	permission.ID = model.ID
	return nil
}

// FindByID retrieves a permission by ID.
func (r *PermissionRepository) FindByID(ctx context.Context, id uint) (*domain.Permission, error) {
	var model PermissionModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindByResourceAndAction retrieves a permission by resource and action.
func (r *PermissionRepository) FindByResourceAndAction(ctx context.Context, resource, action string) (*domain.Permission, error) {
	var model PermissionModel
	err := r.db.WithContext(ctx).Where("resource = ? AND action = ?", resource, action).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, err
	}

	return model.ToDomain()
}

// Update saves permission changes.
func (r *PermissionRepository) Update(ctx context.Context, permission *domain.Permission) error {
	model := &PermissionModel{}
	model.FromDomain(permission)

	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a permission by ID.
func (r *PermissionRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&PermissionModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("permission not found")
	}
	return nil
}

// List retrieves all permissions.
func (r *PermissionRepository) List(ctx context.Context) ([]*domain.Permission, error) {
	var models []PermissionModel
	err := r.db.WithContext(ctx).Find(&models).Error
	if err != nil {
		return nil, err
	}

	permissions := make([]*domain.Permission, len(models))
	for i, model := range models {
		permission, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		permissions[i] = permission
	}

	return permissions, nil
}

// FindByResource retrieves permissions by resource.
func (r *PermissionRepository) FindByResource(ctx context.Context, resource string) ([]*domain.Permission, error) {
	var models []PermissionModel
	err := r.db.WithContext(ctx).Where("resource = ?", resource).Find(&models).Error
	if err != nil {
		return nil, err
	}

	permissions := make([]*domain.Permission, len(models))
	for i, model := range models {
		permission, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		permissions[i] = permission
	}

	return permissions, nil
}

// Count returns the total number of permissions.
func (r *PermissionRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&PermissionModel{}).Count(&count).Error
	return count, err
}

// ExistsByResourceAndAction checks if a permission exists with the given resource and action.
func (r *PermissionRepository) ExistsByResourceAndAction(ctx context.Context, resource, action string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&PermissionModel{}).Where("resource = ? AND action = ?", resource, action).Count(&count).Error
	return count > 0, err
}

// ExistsByID checks if a permission exists with the given ID.
func (r *PermissionRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&PermissionModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Ensure repository implements its interface
var _ repository.PermissionRepository = (*PermissionRepository)(nil)
