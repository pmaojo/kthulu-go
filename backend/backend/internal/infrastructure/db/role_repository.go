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

// RoleModel represents the database model for roles
// Includes a many-to-many relationship with permissions.
type RoleModel struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`

	// Many-to-many relationship with permissions
	Permissions []PermissionModel `gorm:"many2many:role_permissions;"`
}

// TableName specifies the table name for RoleModel
func (RoleModel) TableName() string {
	return "roles"
}

// ToDomain converts RoleModel to domain.Role
func (r *RoleModel) ToDomain() (*domain.Role, error) {
	role := &domain.Role{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	if len(r.Permissions) > 0 {
		permissions := make([]domain.Permission, len(r.Permissions))
		for i, perm := range r.Permissions {
			domainPerm, err := perm.ToDomain()
			if err != nil {
				return nil, err
			}
			permissions[i] = *domainPerm
		}
		role.Permissions = permissions
	}

	return role, nil
}

// FromDomain converts domain.Role to RoleModel
func (r *RoleModel) FromDomain(role *domain.Role) {
	r.ID = role.ID
	r.Name = role.Name
	r.Description = role.Description
	r.CreatedAt = role.CreatedAt
	r.UpdatedAt = role.UpdatedAt
}

// RoleRepository provides a database-backed implementation of repository.RoleRepository.
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new instance bound to a Gorm database.
func NewRoleRepository(db *gorm.DB) repository.RoleRepository {
	return &RoleRepository{db: db}
}

// Create persists a new role.
func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	model := &RoleModel{}
	model.FromDomain(role)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	role.ID = model.ID
	return nil
}

// FindByID retrieves a role by ID.
func (r *RoleRepository) FindByID(ctx context.Context, id uint) (*domain.Role, error) {
	var model RoleModel
	err := r.db.WithContext(ctx).Preload("Permissions").Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindByUserID retrieves a role associated with a user ID.
func (r *RoleRepository) FindByUserID(ctx context.Context, userID uint) (*domain.Role, error) {
	var user UserModel
	err := r.db.WithContext(ctx).Select("role_id").Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}
	return r.FindByID(ctx, user.RoleID)
}

// FindByName retrieves a role by name.
func (r *RoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	var model RoleModel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// Update saves role changes.
func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) error {
	model := &RoleModel{}
	model.FromDomain(role)

	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a role by ID.
func (r *RoleRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&RoleModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrRoleNotFound
	}
	return nil
}

// List retrieves all roles.
func (r *RoleRepository) List(ctx context.Context) ([]*domain.Role, error) {
	var models []RoleModel
	err := r.db.WithContext(ctx).Preload("Permissions").Find(&models).Error
	if err != nil {
		return nil, err
	}

	roles := make([]*domain.Role, len(models))
	for i, model := range models {
		role, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		roles[i] = role
	}

	return roles, nil
}

// Count returns the total number of roles.
func (r *RoleRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RoleModel{}).Count(&count).Error
	return count, err
}

// ExistsByName checks if a role exists with the given name.
func (r *RoleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RoleModel{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// ExistsByID checks if a role exists with the given ID.
func (r *RoleRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RoleModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// AddPermission adds a permission to a role.
func (r *RoleRepository) AddPermission(ctx context.Context, roleID, permissionID uint) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		roleID, permissionID,
	).Error
}

// RemovePermission removes a permission from a role.
func (r *RoleRepository) RemovePermission(ctx context.Context, roleID, permissionID uint) error {
	return r.db.WithContext(ctx).Exec(
		"DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?",
		roleID, permissionID,
	).Error
}

// GetRolePermissions retrieves all permissions for a role.
func (r *RoleRepository) GetRolePermissions(ctx context.Context, roleID uint) ([]*domain.Permission, error) {
	var models []PermissionModel
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&models).Error
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

// Ensure repositories implement their interfaces
var _ repository.RoleRepository = (*RoleRepository)(nil)
