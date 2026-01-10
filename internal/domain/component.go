// @kthulu:module:components
package domain

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrInvalidComponentType      = errors.New("invalid component type")
	ErrComponentGenerationFailed = errors.New("component generation failed")
)

// ComponentRequest represents a request to generate a component
type ComponentRequest struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	Module        string `json:"module,omitempty"`
	WithTests     *bool  `json:"withTests,omitempty"`
	WithMigration *bool  `json:"withMigration,omitempty"`
	Fields        string `json:"fields,omitempty"`
	Relations     string `json:"relations,omitempty"`
	Force         *bool  `json:"force,omitempty"`
	ProjectPath   string `json:"projectPath"`
}

// ComponentGenerationResult represents the result of component generation
type ComponentGenerationResult struct {
	Success     bool      `json:"success"`
	Files       []string  `json:"files,omitempty"`
	Errors      []string  `json:"errors,omitempty"`
	Warnings    []string  `json:"warnings,omitempty"`
	GeneratedAt time.Time `json:"generatedAt"`
}

// NewComponentRequest validates and creates a component request
func NewComponentRequest(req ComponentRequest) (*ComponentRequest, error) {
	if req.Type == "" {
		return nil, ErrInvalidComponentType
	}
	if req.Name == "" {
		return nil, errors.New("component name is required")
	}
	if req.ProjectPath == "" {
		return nil, errors.New("project path is required")
	}

	// Set defaults
	if req.WithTests == nil {
		req.WithTests = &[]bool{true}[0]
	}
	if req.WithMigration == nil {
		req.WithMigration = &[]bool{false}[0]
	}
	if req.Force == nil {
		req.Force = &[]bool{false}[0]
	}

	return &req, nil
}

// ComponentTemplate represents a template for component generation
type ComponentTemplate struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Type        string    `json:"type" gorm:"not null"`
	Name        string    `json:"name" gorm:"not null"`
	Language    string    `json:"language" gorm:"not null"`
	Framework   string    `json:"framework,omitempty"`
	Template    string    `json:"template" gorm:"type:text"` // template content
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NewComponentTemplate creates a new component template
func NewComponentTemplate(componentType, name, language, template string) (*ComponentTemplate, error) {
	if componentType == "" || name == "" || language == "" || template == "" {
		return nil, errors.New("type, name, language, and template are required")
	}

	now := time.Now()
	compTemplate := &ComponentTemplate{
		Type:      componentType,
		Name:      name,
		Language:  language,
		Template:  template,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return compTemplate, nil
}

// Update updates component template fields
func (c *ComponentTemplate) Update(updates ComponentTemplate) {
	if updates.Framework != "" {
		c.Framework = updates.Framework
	}
	if updates.Template != "" {
		c.Template = updates.Template
	}
	if updates.Description != "" {
		c.Description = updates.Description
	}
	c.UpdatedAt = time.Now()
}
