// @kthulu:module:access
package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/kthulu/kthulu-go/backend/core"
	"github.com/kthulu/kthulu-go/backend/internal/domain"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
)

// AccessUseCase orchestrates role and permission management workflows.
type AccessUseCase struct {
	roles       repository.RoleRepository
	permissions repository.PermissionRepository
	users       repository.UserRepository
	logger      core.Logger
}

// NewAccessUseCase builds an AccessUseCase instance.
func NewAccessUseCase(
	roles repository.RoleRepository,
	permissions repository.PermissionRepository,
	users repository.UserRepository,
	logger core.Logger,
) *AccessUseCase {
	return &AccessUseCase{
		roles:       roles,
		permissions: permissions,
		users:       users,
		logger:      logger,
	}
}

// CreateRoleRequest contains the data needed to create a new role
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

// CreateRole creates a new role.
func (a *AccessUseCase) CreateRole(ctx context.Context, req CreateRoleRequest) (*domain.Role, error) {
	a.logger.Info("Create role request", "name", req.Name)

	// Check if role already exists
	exists, err := a.roles.ExistsByName(ctx, req.Name)
	if err != nil {
		a.logger.Error("Failed to check role existence", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to check role existence: %w", err)
	}
	if exists {
		a.logger.Warn("Role creation attempted for existing name", "name", req.Name)
		return nil, domain.ErrRoleAlreadyExists
	}

	// Create role domain entity
	role, err := domain.NewRole(req.Name, req.Description)
	if err != nil {
		a.logger.Error("Failed to create role domain entity", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Persist role
	if err := a.roles.Create(ctx, role); err != nil {
		a.logger.Error("Failed to persist role", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	a.logger.Info("Role created successfully", "roleId", role.ID, "name", req.Name)
	return role, nil
}

// GetRole retrieves a role by ID.
func (a *AccessUseCase) GetRole(ctx context.Context, roleID uint) (*domain.Role, error) {
	a.logger.Info("Get role request", "roleId", roleID)

	role, err := a.roles.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, domain.ErrRoleNotFound) {
			a.logger.Warn("Role not found", "roleId", roleID)
			return nil, domain.ErrRoleNotFound
		}
		a.logger.Error("Failed to find role", "roleId", roleID, "error", err)
		return nil, fmt.Errorf("failed to find role: %w", err)
	}

	a.logger.Info("Role retrieved successfully", "roleId", roleID)
	return role, nil
}

// ListRoles retrieves all roles.
func (a *AccessUseCase) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	a.logger.Info("List roles request")

	roles, err := a.roles.List(ctx)
	if err != nil {
		a.logger.Error("Failed to list roles", "error", err)
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	a.logger.Info("Roles listed successfully", "count", len(roles))
	return roles, nil
}

// UpdateRoleRequest contains the data needed to update a role
type UpdateRoleRequest struct {
	Description *string `json:"description,omitempty"`
}

// UpdateRole updates a role.
func (a *AccessUseCase) UpdateRole(ctx context.Context, roleID uint, req UpdateRoleRequest) (*domain.Role, error) {
	a.logger.Info("Update role request", "roleId", roleID)

	// Find role
	role, err := a.roles.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, domain.ErrRoleNotFound) {
			a.logger.Warn("Update attempted for non-existent role", "roleId", roleID)
			return nil, domain.ErrRoleNotFound
		}
		a.logger.Error("Failed to find role for update", "roleId", roleID, "error", err)
		return nil, fmt.Errorf("failed to find role: %w", err)
	}

	// Update description if provided
	if req.Description != nil {
		role.UpdateDescription(*req.Description)
	}

	// Save changes
	if err := a.roles.Update(ctx, role); err != nil {
		a.logger.Error("Failed to update role", "roleId", roleID, "error", err)
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	a.logger.Info("Role updated successfully", "roleId", roleID)
	return role, nil
}

// DeleteRole deletes a role.
func (a *AccessUseCase) DeleteRole(ctx context.Context, roleID uint) error {
	a.logger.Info("Delete role request", "roleId", roleID)

	// Check if role exists
	exists, err := a.roles.ExistsByID(ctx, roleID)
	if err != nil {
		a.logger.Error("Failed to check role existence for deletion", "roleId", roleID, "error", err)
		return fmt.Errorf("failed to check role existence: %w", err)
	}
	if !exists {
		a.logger.Warn("Delete attempted for non-existent role", "roleId", roleID)
		return domain.ErrRoleNotFound
	}

	// Delete role
	if err := a.roles.Delete(ctx, roleID); err != nil {
		a.logger.Error("Failed to delete role", "roleId", roleID, "error", err)
		return fmt.Errorf("failed to delete role: %w", err)
	}

	a.logger.Info("Role deleted successfully", "roleId", roleID)
	return nil
}

// CreatePermissionRequest contains the data needed to create a new permission
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Resource    string `json:"resource" validate:"required"`
	Action      string `json:"action" validate:"required"`
}

// CreatePermission creates a new permission.
func (a *AccessUseCase) CreatePermission(ctx context.Context, req CreatePermissionRequest) (*domain.Permission, error) {
	a.logger.Info("Create permission request", "name", req.Name, "resource", req.Resource, "action", req.Action)

	// Check if permission already exists
	exists, err := a.permissions.ExistsByResourceAndAction(ctx, req.Resource, req.Action)
	if err != nil {
		a.logger.Error("Failed to check permission existence", "resource", req.Resource, "action", req.Action, "error", err)
		return nil, fmt.Errorf("failed to check permission existence: %w", err)
	}
	if exists {
		a.logger.Warn("Permission creation attempted for existing resource/action", "resource", req.Resource, "action", req.Action)
		return nil, errors.New("permission already exists")
	}

	// Create permission domain entity
	permission, err := domain.NewPermission(req.Name, req.Description, req.Resource, req.Action)
	if err != nil {
		a.logger.Error("Failed to create permission domain entity", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	// Persist permission
	if err := a.permissions.Create(ctx, permission); err != nil {
		a.logger.Error("Failed to persist permission", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	a.logger.Info("Permission created successfully", "permissionId", permission.ID, "name", req.Name)
	return permission, nil
}

// ListPermissions retrieves all permissions.
func (a *AccessUseCase) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	a.logger.Info("List permissions request")

	permissions, err := a.permissions.List(ctx)
	if err != nil {
		a.logger.Error("Failed to list permissions", "error", err)
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	a.logger.Info("Permissions listed successfully", "count", len(permissions))
	return permissions, nil
}

// AssignPermissionToRoleRequest contains the data needed to assign a permission to a role
type AssignPermissionToRoleRequest struct {
	RoleID       uint `json:"roleId" validate:"required"`
	PermissionID uint `json:"permissionId" validate:"required"`
}

// AssignPermissionToRole assigns a permission to a role.
func (a *AccessUseCase) AssignPermissionToRole(ctx context.Context, req AssignPermissionToRoleRequest) error {
	a.logger.Info("Assign permission to role request", "roleId", req.RoleID, "permissionId", req.PermissionID)

	// Check if role exists
	roleExists, err := a.roles.ExistsByID(ctx, req.RoleID)
	if err != nil {
		a.logger.Error("Failed to check role existence", "roleId", req.RoleID, "error", err)
		return fmt.Errorf("failed to check role existence: %w", err)
	}
	if !roleExists {
		a.logger.Warn("Permission assignment attempted for non-existent role", "roleId", req.RoleID)
		return domain.ErrRoleNotFound
	}

	// Check if permission exists
	permissionExists, err := a.permissions.ExistsByID(ctx, req.PermissionID)
	if err != nil {
		a.logger.Error("Failed to check permission existence", "permissionId", req.PermissionID, "error", err)
		return fmt.Errorf("failed to check permission existence: %w", err)
	}
	if !permissionExists {
		a.logger.Warn("Permission assignment attempted for non-existent permission", "permissionId", req.PermissionID)
		return errors.New("permission not found")
	}

	// Assign permission to role
	if err := a.roles.AddPermission(ctx, req.RoleID, req.PermissionID); err != nil {
		a.logger.Error("Failed to assign permission to role", "roleId", req.RoleID, "permissionId", req.PermissionID, "error", err)
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	a.logger.Info("Permission assigned to role successfully", "roleId", req.RoleID, "permissionId", req.PermissionID)
	return nil
}

// RemovePermissionFromRole removes a permission from a role.
func (a *AccessUseCase) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint) error {
	a.logger.Info("Remove permission from role request", "roleId", roleID, "permissionId", permissionID)

	// Remove permission from role
	if err := a.roles.RemovePermission(ctx, roleID, permissionID); err != nil {
		a.logger.Error("Failed to remove permission from role", "roleId", roleID, "permissionId", permissionID, "error", err)
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	a.logger.Info("Permission removed from role successfully", "roleId", roleID, "permissionId", permissionID)
	return nil
}

// CheckUserPermission checks if a user has a specific permission.
func (a *AccessUseCase) CheckUserPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	a.logger.Info("Check user permission request", "userId", userID, "resource", resource, "action", action)

	// Find user
	user, err := a.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			a.logger.Warn("Permission check for non-existent user", "userId", userID)
			return false, domain.ErrUserNotFound
		}
		a.logger.Error("Failed to find user for permission check", "userId", userID, "error", err)
		return false, fmt.Errorf("failed to find user: %w", err)
	}

	// Find user's role
	role, err := a.roles.FindByID(ctx, user.RoleID)
	if err != nil {
		if errors.Is(err, domain.ErrRoleNotFound) {
			a.logger.Warn("Permission check for user with non-existent role", "userId", userID, "roleId", user.RoleID)
			return false, domain.ErrRoleNotFound
		}
		a.logger.Error("Failed to find user role for permission check", "userId", userID, "roleId", user.RoleID, "error", err)
		return false, fmt.Errorf("failed to find user role: %w", err)
	}

	// Check if role has permission
	hasPermission := role.HasPermission(resource, action)

	a.logger.Info("User permission check completed", "userId", userID, "resource", resource, "action", action, "hasPermission", hasPermission)
	return hasPermission, nil
}
