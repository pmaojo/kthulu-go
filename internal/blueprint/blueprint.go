package blueprint

import (
	"time"
)

// Requirement represents a high-level project requirement
type Requirement struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
	Priority    string `yaml:"priority"` // High, Medium, Low
	Status      string `yaml:"status"`   // Pending, InProgress, Done
	Created     string `yaml:"created"`
}

// ModuleConfig represents detailed configuration for a module
type ModuleConfig struct {
	Fields []string `yaml:"fields,omitempty"`
}

// ProjectBlueprint represents the desired state of a project
type ProjectBlueprint struct {
	Name         string                  `yaml:"name"`
	Description  string                  `yaml:"description"`
	Template     string                  `yaml:"template"`
	Features     []string                `yaml:"features"`
	Modules      map[string]ModuleConfig `yaml:"modules,omitempty"`
	Database     string                  `yaml:"database"`
	Frontend     string                  `yaml:"frontend"`
	Auth         string                  `yaml:"auth"`
	Requirements []Requirement           `yaml:"requirements,omitempty"`
}

// NewRequirement creates a new requirement with default status and timestamp
func NewRequirement(title, description, priority, id string) Requirement {
	return Requirement{
		ID:          id,
		Title:       title,
		Description: description,
		Priority:    priority,
		Status:      "Pending",
		Created:     time.Now().Format(time.RFC3339),
	}
}
