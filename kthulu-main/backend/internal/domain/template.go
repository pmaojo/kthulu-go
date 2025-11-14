// @kthulu:module:templates
package domain

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrInvalidTemplateName = errors.New("invalid template name")
	ErrTemplateNotFound    = errors.New("template not found")
)

// Template represents a code template
type Template struct {
	ID          uint              `json:"id" gorm:"primaryKey"`
	Name        string            `json:"name" gorm:"not null;uniqueIndex"`
	Version     string            `json:"version,omitempty"`
	Description string            `json:"description,omitempty"`
	Author      string            `json:"author,omitempty"`
	Category    string            `json:"category,omitempty"`
	Tags        []string          `json:"tags" gorm:"serializer:json"`
	Content     map[string]string `json:"content" gorm:"serializer:json"` // file path -> content
	Remote      bool              `json:"remote"`
	URL         string            `json:"url,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// TemplateInfo represents template metadata
type TemplateInfo struct {
	Name          string   `json:"name"`
	Version       string   `json:"version,omitempty"`
	LatestVersion string   `json:"latest_version,omitempty"`
	Description   string   `json:"description,omitempty"`
	Author        string   `json:"author,omitempty"`
	Category      string   `json:"category,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Remote        bool     `json:"remote"`
	URL           string   `json:"url,omitempty"`
}

// TemplateRenderRequest represents a request to render a template
type TemplateRenderRequest struct {
	Name string                 `json:"name"`
	Vars map[string]interface{} `json:"vars,omitempty"`
}

// TemplateRenderResult represents the result of template rendering
type TemplateRenderResult struct {
	Files map[string]string `json:"files"` // base64 encoded
}

// TemplateSyncResult represents the result of template synchronization
type TemplateSyncResult struct {
	Source              string `json:"source"`
	Destination         string `json:"destination"`
	FilesCopied         int    `json:"filesCopied"`
	TemplatesRegistered int    `json:"templatesRegistered"`
	ManifestPath        string `json:"manifestPath"`
}

// TemplateDriftReport represents template drift detection results
type TemplateDriftReport struct {
	Added   []string              `json:"added"`
	Removed []string              `json:"removed"`
	Changed []TemplateDriftChange `json:"changed"`
}

// TemplateDriftChange represents a change in template
type TemplateDriftChange struct {
	Path             string `json:"path"`
	ExpectedChecksum string `json:"expectedChecksum"`
	ActualChecksum   string `json:"actualChecksum"`
}

// TemplateRegistry represents a template registry source
type TemplateRegistry struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null;uniqueIndex"`
	URL       string    `json:"url" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewTemplate creates a new template with validation
func NewTemplate(name string) (*Template, error) {
	if name == "" {
		return nil, ErrInvalidTemplateName
	}

	now := time.Now()
	template := &Template{
		Name:      name,
		Content:   make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}

	return template, nil
}

// Update updates template fields
func (t *Template) Update(updates Template) {
	if updates.Version != "" {
		t.Version = updates.Version
	}
	if updates.Description != "" {
		t.Description = updates.Description
	}
	if updates.Author != "" {
		t.Author = updates.Author
	}
	if updates.Category != "" {
		t.Category = updates.Category
	}
	if updates.Tags != nil {
		t.Tags = updates.Tags
	}
	if updates.Content != nil {
		t.Content = updates.Content
	}
	t.Remote = updates.Remote
	if updates.URL != "" {
		t.URL = updates.URL
	}
	t.UpdatedAt = time.Now()
}

// AddFile adds a file to the template content
func (t *Template) AddFile(path, content string) {
	if t.Content == nil {
		t.Content = make(map[string]string)
	}
	t.Content[path] = content
	t.UpdatedAt = time.Now()
}

// RemoveFile removes a file from the template content
func (t *Template) RemoveFile(path string) {
	if t.Content != nil {
		delete(t.Content, path)
		t.UpdatedAt = time.Now()
	}
}

// GetFile retrieves a file from the template content
func (t *Template) GetFile(path string) (string, bool) {
	if t.Content == nil {
		return "", false
	}
	content, exists := t.Content[path]
	return content, exists
}

// GetAllFiles returns all files in the template
func (t *Template) GetAllFiles() map[string]string {
	if t.Content == nil {
		return make(map[string]string)
	}
	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range t.Content {
		result[k] = v
	}
	return result
}

// NewTemplateRegistry creates a new template registry
func NewTemplateRegistry(name, url string) (*TemplateRegistry, error) {
	if name == "" || url == "" {
		return nil, errors.New("name and url are required")
	}

	now := time.Now()
	registry := &TemplateRegistry{
		Name:      name,
		URL:       url,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return registry, nil
}
