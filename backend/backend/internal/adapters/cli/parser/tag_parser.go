package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TagParser parses Go code to extract Kthulu tags
type TagParser struct {
	cache Cache
}

// NewTagParser creates a new tag parser with optional cache
func NewTagParser(cache Cache) *TagParser {
	if cache == nil {
		cache = NewNullCache() // Simple implementation that doesn't cache
	}
	return &TagParser{cache: cache}
}

// AnalyzeProject analyzes a complete project for Kthulu tags
func (p *TagParser) AnalyzeProject(projectPath string) (*ProjectAnalysis, error) {
	analysis := &ProjectAnalysis{
		ProjectPath:  projectPath,
		Modules:      make(map[string]*Module),
		Dependencies: []Dependency{},
		Tags:         []Tag{},
		LastScanned:  time.Now(),
	}

	// Walk through the project directory
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor and node_modules directories
		if strings.Contains(path, "vendor/") || strings.Contains(path, "node_modules/") {
			return nil
		}

		// Parse the file
		fileAnalysis, err := p.AnalyzeFile(path)
		if err != nil {
			// Continue with other files if one fails
			fmt.Printf("Warning: Could not parse %s: %v\n", path, err)
			return nil
		}

		// Merge file analysis into project analysis
		p.mergeFileAnalysis(analysis, fileAnalysis)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking project directory: %w", err)
	}

	return analysis, nil
}

// AnalyzeFile analyzes a single Go file for Kthulu tags
func (p *TagParser) AnalyzeFile(filePath string) (*FileAnalysis, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing Go file: %w", err)
	}

	analysis := &FileAnalysis{
		FilePath: filePath,
		Package:  node.Name.Name,
		Tags:     []Tag{},
		Imports:  []string{},
		Symbols:  []Symbol{},
	}

	// Extract imports
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		analysis.Imports = append(analysis.Imports, importPath)
	}

	// Extract comments for tags
	for _, commentGroup := range node.Comments {
		for _, comment := range commentGroup.List {
			if tag := p.parseTag(comment.Text); tag != nil {
				analysis.Tags = append(analysis.Tags, *tag)
			}
		}
	}

	// Walk the AST to find symbols
	ast.Inspect(node, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name.IsExported() {
				symbol := Symbol{
					Name: node.Name.Name,
					Type: "function",
					Line: fset.Position(node.Pos()).Line,
				}
				analysis.Symbols = append(analysis.Symbols, symbol)
			}
		case *ast.TypeSpec:
			if node.Name.IsExported() {
				symbol := Symbol{
					Name: node.Name.Name,
					Type: "type",
					Line: fset.Position(node.Pos()).Line,
				}
				analysis.Symbols = append(analysis.Symbols, symbol)
			}
		}
		return true
	})

	return analysis, nil
}

// parseTag extracts a Kthulu tag from a comment
func (p *TagParser) parseTag(comment string) *Tag {
	// Remove comment markers
	comment = strings.TrimPrefix(comment, "//")
	comment = strings.TrimPrefix(comment, "/*")
	comment = strings.TrimSuffix(comment, "*/")
	comment = strings.TrimSpace(comment)

	// Check if it's a Kthulu tag
	if !strings.HasPrefix(comment, "@kthulu:") {
		return nil
	}

	// Parse the tag
	tagContent := strings.TrimPrefix(comment, "@kthulu:")
	parts := strings.SplitN(tagContent, ":", 2)

	tag := &Tag{
		Type:       TagType(parts[0]),
		Content:    comment,
		Attributes: make(map[string]string),
	}

	if len(parts) > 1 {
		tag.Value = parts[1]
	}

	return tag
}

// mergeFileAnalysis merges file analysis into project analysis
func (p *TagParser) mergeFileAnalysis(project *ProjectAnalysis, file *FileAnalysis) {
	// Add tags
	project.Tags = append(project.Tags, file.Tags...)

	// Process module tags to build module map
	for _, tag := range file.Tags {
		switch tag.Type {
		case TagTypeModule:
			moduleName := tag.Value
			if moduleName != "" {
				module, exists := project.Modules[moduleName]
				if !exists {
					module = &Module{
						Name:         moduleName,
						Package:      file.Package,
						Files:        []string{},
						Dependencies: []string{},
						Tags:         []Tag{},
					}
					project.Modules[moduleName] = module
				}
				module.Files = append(module.Files, file.FilePath)
				module.Tags = append(module.Tags, tag)
			}
		case TagTypeDependency:
			// Add dependency
			dependency := Dependency{
				From: file.Package,
				To:   tag.Value,
				Type: "module",
			}
			project.Dependencies = append(project.Dependencies, dependency)
		}
	}
}

// NullCache is a cache implementation that doesn't actually cache
type NullCache struct{}

func NewNullCache() Cache {
	return &NullCache{}
}

func (c *NullCache) Get(key string) ([]byte, bool) {
	return nil, false
}

func (c *NullCache) Set(key string, value []byte) error {
	return nil
}

func (c *NullCache) Delete(key string) error {
	return nil
}

func (c *NullCache) Clear() error {
	return nil
}
