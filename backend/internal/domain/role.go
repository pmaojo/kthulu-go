// @kthulu:module:access
package domain

import (
	"errors"
	"strings"
	"time"
)

// Role-related errors
var (
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleAlreadyExists = errors.New("role already exists")
	ErrInvalidRoleName   = errors.New("invalid role name")
)

// Predefined role names
const (
	RoleAdmin     = "admin"
	RoleUser      = "user"
	RoleModerator = "moderator"
)

// Role represents a user role with permissions
type Role struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// Permission represents a specific permission
type Permission struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NewRole creates a new role with validation
func NewRole(name, description string) (*Role, error) {
	name = strings.TrimSpace(strings.ToLower(name))

	if name == "" {
		return nil, ErrInvalidRoleName
	}

	// Validate role name format (alphanumeric and underscores only)
	if !isValidRoleName(name) {
		return nil, ErrInvalidRoleName
	}

	now := time.Now()
	role := &Role{
		Name:        name,
		Description: strings.TrimSpace(description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return role, nil
}

// NewPermission creates a new permission with validation
func NewPermission(name, description, resource, action string) (*Permission, error) {
	name = strings.TrimSpace(name)
	resource = strings.TrimSpace(strings.ToLower(resource))
	action = strings.TrimSpace(strings.ToLower(action))

	if name == "" || resource == "" || action == "" {
		return nil, errors.New("permission name, resource, and action are required")
	}

	now := time.Now()
	permission := &Permission{
		Name:        name,
		Description: strings.TrimSpace(description),
		Resource:    resource,
		Action:      action,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return permission, nil
}

// IsAdmin returns true if the role is an admin role
func (r *Role) IsAdmin() bool {
	return r.Name == RoleAdmin
}

// IsModerator returns true if the role is a moderator role
func (r *Role) IsModerator() bool {
	return r.Name == RoleModerator
}

// IsUser returns true if the role is a regular user role
func (r *Role) IsUser() bool {
	return r.Name == RoleUser
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(resource, action string) bool {
	// Admin has all permissions
	if r.IsAdmin() {
		return true
	}

	for _, permission := range r.Permissions {
		if permission.Resource == resource && permission.Action == action {
			return true
		}
	}

	return false
}

// AddPermission adds a permission to the role
func (r *Role) AddPermission(permission Permission) {
	// Check if permission already exists
	for _, p := range r.Permissions {
		if p.Resource == permission.Resource && p.Action == permission.Action {
			return // Permission already exists
		}
	}

	r.Permissions = append(r.Permissions, permission)
	r.UpdatedAt = time.Now()
}

// RemovePermission removes a permission from the role
func (r *Role) RemovePermission(resource, action string) {
	for i, permission := range r.Permissions {
		if permission.Resource == resource && permission.Action == action {
			r.Permissions = append(r.Permissions[:i], r.Permissions[i+1:]...)
			r.UpdatedAt = time.Now()
			break
		}
	}
}

// UpdateDescription updates the role description
func (r *Role) UpdateDescription(description string) {
	r.Description = strings.TrimSpace(description)
	r.UpdatedAt = time.Now()
}

// isValidRoleName validates role name format
func isValidRoleName(name string) bool {
	if len(name) < 2 || len(name) > 50 {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}
