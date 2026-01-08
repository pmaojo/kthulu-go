// @kthulu:module:projects
package domain

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrInvalidProjectName = errors.New("invalid project name")
	ErrProjectNotFound    = errors.New("project not found")
	ErrInvalidModules     = errors.New("invalid modules configuration")
)

// Project represents a code generation project
type Project struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;uniqueIndex"`
	Modules     []string  `json:"modules" gorm:"serializer:json"`
	Template    string    `json:"template,omitempty"`
	Database    string    `json:"database,omitempty"`
	Frontend    string    `json:"frontend,omitempty"`
	SkipGit     bool      `json:"skipGit"`
	SkipDocker  bool      `json:"skipDocker"`
	Author      string    `json:"author,omitempty"`
	License     string    `json:"license,omitempty"`
	Description string    `json:"description,omitempty"`
	Path        string    `json:"path,omitempty"`
	DryRun      bool      `json:"dryRun"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ProjectRequest represents a request to create or plan a project
type ProjectRequest struct {
	Name        string   `json:"name"`
	Modules     []string `json:"modules,omitempty"`
	Template    string   `json:"template,omitempty"`
	Database    string   `json:"database,omitempty"`
	Frontend    string   `json:"frontend,omitempty"`
	SkipGit     *bool    `json:"skipGit,omitempty"`
	SkipDocker  *bool    `json:"skipDocker,omitempty"`
	Author      string   `json:"author,omitempty"`
	License     string   `json:"license,omitempty"`
	Description string   `json:"description,omitempty"`
	Path        string   `json:"path,omitempty"`
	DryRun      *bool    `json:"dryRun,omitempty"`
}

// ProjectStructure represents the structure of a generated project
type ProjectStructure struct {
	Name        string          `json:"name"`
	Path        string          `json:"path"`
	Backend     *BackendConfig  `json:"backend,omitempty"`
	Frontend    *FrontendConfig `json:"frontend,omitempty"`
	Database    *DatabaseConfig `json:"database,omitempty"`
	Docker      *DockerConfig   `json:"docker,omitempty"`
	Config      map[string]any  `json:"config,omitempty"`
	Modules     []ModuleInfo    `json:"modules,omitempty"`
	Author      string          `json:"author,omitempty"`
	License     string          `json:"license,omitempty"`
	Description string          `json:"description,omitempty"`
}

// BackendConfig represents backend configuration
type BackendConfig struct {
	PackageName  string   `json:"packageName"`
	Modules      []string `json:"modules"`
	Architecture string   `json:"architecture"`
	Template     string   `json:"template"`
}

// FrontendConfig represents frontend configuration
type FrontendConfig struct {
	Framework string   `json:"framework"`
	Language  string   `json:"language"`
	Modules   []string `json:"modules"`
	Template  string   `json:"template"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type       string   `json:"type"`
	Migrations []string `json:"migrations"`
}

// DockerConfig represents Docker configuration
type DockerConfig struct {
	Enabled  bool     `json:"enabled"`
	Services []string `json:"services"`
}

// ProjectPlan represents a complete project plan
type ProjectPlan struct {
	Options                 ProjectRequest   `json:"options"`
	Structure               ProjectStructure `json:"structure"`
	Modules                 []string         `json:"modules"`
	ProjectDirectories      []string         `json:"projectDirectories"`
	BackendTemplate         string           `json:"backendTemplate,omitempty"`
	BackendTemplateVersion  string           `json:"backendTemplateVersion,omitempty"`
	BackendFiles            []string         `json:"backendFiles,omitempty"`
	FrontendTemplate        string           `json:"frontendTemplate,omitempty"`
	FrontendTemplateVersion string           `json:"frontendTemplateVersion,omitempty"`
	FrontendFiles           []string         `json:"frontendFiles,omitempty"`
	StaticFiles             []string         `json:"staticFiles,omitempty"`
	ConfigFiles             []string         `json:"configFiles,omitempty"`
	MigrationFiles          []string         `json:"migrationFiles,omitempty"`
	DockerServices          []string         `json:"dockerServices,omitempty"`
}

// NewProject creates a new project with validation
func NewProject(req ProjectRequest) (*Project, error) {
	if req.Name == "" {
		return nil, ErrInvalidProjectName
	}

	now := time.Now()
	project := &Project{
		Name:        req.Name,
		Modules:     req.Modules,
		Template:    req.Template,
		Database:    req.Database,
		Frontend:    req.Frontend,
		SkipGit:     req.SkipGit != nil && *req.SkipGit,
		SkipDocker:  req.SkipDocker != nil && *req.SkipDocker,
		Author:      req.Author,
		License:     req.License,
		Description: req.Description,
		Path:        req.Path,
		DryRun:      req.DryRun != nil && *req.DryRun,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return project, nil
}

// Update updates project fields
func (p *Project) Update(req ProjectRequest) {
	if req.Modules != nil {
		p.Modules = req.Modules
	}
	if req.Template != "" {
		p.Template = req.Template
	}
	if req.Database != "" {
		p.Database = req.Database
	}
	if req.Frontend != "" {
		p.Frontend = req.Frontend
	}
	if req.SkipGit != nil {
		p.SkipGit = *req.SkipGit
	}
	if req.SkipDocker != nil {
		p.SkipDocker = *req.SkipDocker
	}
	if req.Author != "" {
		p.Author = req.Author
	}
	if req.License != "" {
		p.License = req.License
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Path != "" {
		p.Path = req.Path
	}
	if req.DryRun != nil {
		p.DryRun = *req.DryRun
	}
	p.UpdatedAt = time.Now()
}
