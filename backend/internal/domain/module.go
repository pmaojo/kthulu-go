// @kthulu:module:modules
package domain

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrInvalidModuleName = errors.New("invalid module name")
	ErrModuleNotFound    = errors.New("module not found")
)

// ModuleInfo represents a code module in the catalog
type ModuleInfo struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"not null;uniqueIndex"`
	Description  string         `json:"description,omitempty"`
	Version      string         `json:"version,omitempty"`
	Dependencies []string       `json:"dependencies" gorm:"serializer:json"`
	Optional     bool           `json:"optional"`
	Category     string         `json:"category,omitempty"`
	Tags         []string       `json:"tags" gorm:"serializer:json"`
	Entities     []any          `json:"entities" gorm:"serializer:json"`
	Routes       []any          `json:"routes" gorm:"serializer:json"`
	Migrations   []string       `json:"migrations" gorm:"serializer:json"`
	Frontend     bool           `json:"frontend"`
	Backend      bool           `json:"backend"`
	Config       map[string]any `json:"config" gorm:"serializer:json"`
	Conflicts    []string       `json:"conflicts" gorm:"serializer:json"`
	MinVersion   string         `json:"minVersion,omitempty"`
	MaxVersion   string         `json:"maxVersion,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

// ModuleValidationResult represents the result of module validation
type ModuleValidationResult struct {
	Valid     bool                      `json:"valid"`
	Missing   []string                  `json:"missing,omitempty"`
	Circular  []CircularDependencyChain `json:"circular,omitempty"`
	Conflicts []ModuleConflict          `json:"conflicts,omitempty"`
	Resolved  []string                  `json:"resolved,omitempty"`
	Warnings  []string                  `json:"warnings,omitempty"`
}

// CircularDependencyChain represents a circular dependency
type CircularDependencyChain struct {
	Chain []string `json:"chain"`
}

// ModuleConflict represents a module conflict
type ModuleConflict struct {
	Module    string   `json:"module"`
	Conflicts []string `json:"conflicts"`
	Reason    string   `json:"reason"`
}

// ModuleInjectionPlan represents a plan for injecting modules
type ModuleInjectionPlan struct {
	RequestedModules []string              `json:"requested_modules"`
	ResolvedModules  []string              `json:"resolved_modules"`
	InjectedModules  []string              `json:"injected_modules"`
	ExecutionOrder   []string              `json:"execution_order"`
	ModuleDetails    map[string]ModuleInfo `json:"module_details"`
	Warnings         []string              `json:"warnings,omitempty"`
	Errors           []string              `json:"errors,omitempty"`
}

// NewModuleInfo creates a new module info with validation
func NewModuleInfo(name string) (*ModuleInfo, error) {
	if name == "" {
		return nil, ErrInvalidModuleName
	}

	now := time.Now()
	module := &ModuleInfo{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return module, nil
}

// Update updates module info fields
func (m *ModuleInfo) Update(updates ModuleInfo) {
	if updates.Description != "" {
		m.Description = updates.Description
	}
	if updates.Version != "" {
		m.Version = updates.Version
	}
	if updates.Dependencies != nil {
		m.Dependencies = updates.Dependencies
	}
	m.Optional = updates.Optional
	if updates.Category != "" {
		m.Category = updates.Category
	}
	if updates.Tags != nil {
		m.Tags = updates.Tags
	}
	if updates.Entities != nil {
		m.Entities = updates.Entities
	}
	if updates.Routes != nil {
		m.Routes = updates.Routes
	}
	if updates.Migrations != nil {
		m.Migrations = updates.Migrations
	}
	m.Frontend = updates.Frontend
	m.Backend = updates.Backend
	if updates.Config != nil {
		m.Config = updates.Config
	}
	if updates.Conflicts != nil {
		m.Conflicts = updates.Conflicts
	}
	if updates.MinVersion != "" {
		m.MinVersion = updates.MinVersion
	}
	if updates.MaxVersion != "" {
		m.MaxVersion = updates.MaxVersion
	}
	m.UpdatedAt = time.Now()
}

// HasDependency checks if this module depends on another
func (m *ModuleInfo) HasDependency(moduleName string) bool {
	for _, dep := range m.Dependencies {
		if dep == moduleName {
			return true
		}
	}
	return false
}

// ConflictsWith checks if this module conflicts with another
func (m *ModuleInfo) ConflictsWith(moduleName string) bool {
	for _, conflict := range m.Conflicts {
		if conflict == moduleName {
			return true
		}
	}
	return false
}

// IsCompatible checks version compatibility
func (m *ModuleInfo) IsCompatible(version string) bool {
	// Simple version check - could be enhanced
	if m.MinVersion != "" && version < m.MinVersion {
		return false
	}
	if m.MaxVersion != "" && version > m.MaxVersion {
		return false
	}
	return true
}
