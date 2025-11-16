package parser

import (
	"time"
)

// TagType represents the type of a Kthulu tag
type TagType string

const (
	TagTypeModule     TagType = "module"
	TagTypeDependency TagType = "dependency"
	TagTypeRequires   TagType = "requires"
	TagTypeProvides   TagType = "provides"
	TagTypeObservable TagType = "observable"
	TagTypeSecurity   TagType = "security"
	TagTypeCompliance TagType = "compliance"
	TagTypeHandler    TagType = "handler"
	TagTypeService    TagType = "service"
	TagTypeRepository TagType = "repository"
	TagTypeDomain     TagType = "domain"
	TagTypeGenerated  TagType = "generated"
	TagTypeProject    TagType = "project"
)

// Tag represents a parsed Kthulu tag
type Tag struct {
	Type       TagType           `json:"type"`
	Value      string            `json:"value"`
	Content    string            `json:"content"`
	Line       int               `json:"line"`
	Attributes map[string]string `json:"attributes"`
}

// ProjectAnalysis contains the complete analysis of a Kthulu project
type ProjectAnalysis struct {
	ProjectPath  string             `json:"project_path"`
	Modules      map[string]*Module `json:"modules"`
	Dependencies []Dependency       `json:"dependencies"`
	Tags         []Tag              `json:"tags"`
	LastScanned  time.Time          `json:"last_scanned"`
}

// Module represents a Kthulu module
type Module struct {
	Name         string   `json:"name"`
	Package      string   `json:"package"`
	Files        []string `json:"files"`
	Dependencies []string `json:"dependencies"`
	Tags         []Tag    `json:"tags"`
}

// Dependency represents a dependency relationship
type Dependency struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
	Line int    `json:"line"`
}

// FileAnalysis contains analysis results for a single file
type FileAnalysis struct {
	FilePath string   `json:"file_path"`
	Package  string   `json:"package"`
	Tags     []Tag    `json:"tags"`
	Imports  []string `json:"imports"`
	Symbols  []Symbol `json:"symbols"`
}

// Symbol represents a symbol found in the code
type Symbol struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Line int    `json:"line"`
}

// Cache interface for caching parsed results
type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte) error
	Delete(key string) error
	Clear() error
}
