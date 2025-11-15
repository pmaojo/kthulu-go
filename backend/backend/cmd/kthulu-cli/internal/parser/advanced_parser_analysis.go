package parser

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// analyzeFileAdvanced performs advanced analysis on a single file
func (p *AdvancedTagParser) analyzeFileAdvanced(filePath string) (*FileAnalysis, error) {
	// Check cache for this file
	cacheKey := p.generateFileCacheKey(filePath)
	if cached, found := p.cache.Get(cacheKey); found {
		var analysis FileAnalysis
		if err := json.Unmarshal(cached, &analysis); err == nil {
			return &analysis, nil
		}
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse AST
	file, err := parser.ParseFile(p.fileSet, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	analysis := &FileAnalysis{
		FilePath: filePath,
		Package:  file.Name.Name,
		Tags:     []Tag{},
		Imports:  []string{},
		Symbols:  []Symbol{},
	}

	// Extract imports
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		analysis.Imports = append(analysis.Imports, importPath)
	}

	// Walk AST for symbols and tags
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name != nil {
				symbol := Symbol{
					Name: node.Name.Name,
					Type: "function",
					Line: p.fileSet.Position(node.Pos()).Line,
				}
				analysis.Symbols = append(analysis.Symbols, symbol)

				// Extract tags from comments
				if node.Doc != nil {
					tags := p.extractAdvancedTagsFromComments(node.Doc.List, symbol.Line)
					analysis.Tags = append(analysis.Tags, tags...)
				}
			}

		case *ast.TypeSpec:
			if node.Name != nil {
				symbol := Symbol{
					Name: node.Name.Name,
					Type: "type",
					Line: p.fileSet.Position(node.Pos()).Line,
				}
				analysis.Symbols = append(analysis.Symbols, symbol)

				// Extract tags from comments
				if node.Doc != nil {
					tags := p.extractAdvancedTagsFromComments(node.Doc.List, symbol.Line)
					analysis.Tags = append(analysis.Tags, tags...)
				}
			}

		case *ast.GenDecl:
			// Handle variable and constant declarations
			if node.Tok == token.VAR || node.Tok == token.CONST {
				for _, spec := range node.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valueSpec.Names {
							symbol := Symbol{
								Name: name.Name,
								Type: strings.ToLower(node.Tok.String()),
								Line: p.fileSet.Position(name.Pos()).Line,
							}
							analysis.Symbols = append(analysis.Symbols, symbol)
						}
					}
				}

				// Extract tags from comments
				if node.Doc != nil {
					tags := p.extractAdvancedTagsFromComments(node.Doc.List, p.fileSet.Position(node.Pos()).Line)
					analysis.Tags = append(analysis.Tags, tags...)
				}
			}
		}
		return true
	})

	// Extract tags from top-level comments
	if file.Doc != nil {
		tags := p.extractAdvancedTagsFromComments(file.Doc.List, 1)
		analysis.Tags = append(analysis.Tags, tags...)
	}

	// Cache results
	if data, err := json.Marshal(analysis); err == nil {
		p.cache.Set(cacheKey, data)
	}

	return analysis, nil
}

// extractAdvancedTagsFromComments extracts tags with advanced attribute parsing
func (p *AdvancedTagParser) extractAdvancedTagsFromComments(comments []*ast.Comment, line int) []Tag {
	var tags []Tag

	tagRegex := regexp.MustCompile(`@kthulu:([a-zA-Z_][a-zA-Z0-9_]*):?([a-zA-Z0-9_]*)?(?:\s+(.+))?`)

	for _, comment := range comments {
		commentText := strings.TrimPrefix(comment.Text, "//")
		commentText = strings.TrimPrefix(commentText, "/*")
		commentText = strings.TrimSuffix(commentText, "*/")
		commentText = strings.TrimSpace(commentText)

		matches := tagRegex.FindStringSubmatch(commentText)
		if len(matches) >= 2 {
			tag := Tag{
				Type:       TagType(matches[1]),
				Value:      "",
				Content:    commentText,
				Line:       line,
				Attributes: make(map[string]string),
			}

			// Extract value if present
			if len(matches) >= 3 && matches[2] != "" {
				tag.Value = matches[2]
			}

			// Parse attributes from the remaining content
			if len(matches) >= 4 && matches[3] != "" {
				attributes := p.parseAdvancedAttributes(matches[3])
				tag.Attributes = attributes
			}

			tags = append(tags, tag)
		}
	}

	return tags
}

// parseAdvancedAttributes parses key=value attributes from tag content
func (p *AdvancedTagParser) parseAdvancedAttributes(content string) map[string]string {
	attributes := make(map[string]string)

	// Support multiple formats:
	// key=value key2=value2
	// key="quoted value" key2='single quoted'
	// key=value,key2=value2

	// Split by spaces and commas, but respect quotes
	parts := p.smartSplit(content, []rune{' ', ',', '\t'})

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Look for key=value pattern
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])

				// Remove quotes if present
				value = strings.Trim(value, `"'`)

				attributes[key] = value
			}
		} else {
			// If no '=' found, treat as boolean flag
			attributes[part] = "true"
		}
	}

	return attributes
}

// smartSplit splits a string by delimiters while respecting quotes
func (p *AdvancedTagParser) smartSplit(s string, delimiters []rune) []string {
	var result []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune

	for _, char := range s {
		if !inQuotes {
			if char == '"' || char == '\'' {
				inQuotes = true
				quoteChar = char
				current.WriteRune(char)
			} else if p.isDelimiter(char, delimiters) {
				if current.Len() > 0 {
					result = append(result, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(char)
			}
		} else {
			current.WriteRune(char)
			if char == quoteChar {
				inQuotes = false
			}
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// isDelimiter checks if a character is in the delimiter list
func (p *AdvancedTagParser) isDelimiter(char rune, delimiters []rune) bool {
	for _, delimiter := range delimiters {
		if char == delimiter {
			return true
		}
	}
	return false
}

// mergeFileAnalysis merges file analysis into project analysis
func (p *AdvancedTagParser) mergeFileAnalysis(project *ProjectAnalysis, file *FileAnalysis) {
	// Group symbols by module based on tags
	moduleName := p.extractModuleName(file.Tags)
	if moduleName == "" {
		moduleName = file.Package
	}

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
	module.Tags = append(module.Tags, file.Tags...)

	// Extract dependencies from imports
	for _, imp := range file.Imports {
		if p.isProjectImport(imp, project.ProjectPath) {
			depModule := p.extractModuleFromImport(imp)
			if depModule != "" && depModule != moduleName {
				module.Dependencies = append(module.Dependencies, depModule)

				// Add to project dependencies
				dependency := Dependency{
					From: moduleName,
					To:   depModule,
					Type: "import",
					Line: 0, // Could be enhanced to track specific line
				}
				project.Dependencies = append(project.Dependencies, dependency)
			}
		}
	}

	// Remove duplicates from dependencies
	module.Dependencies = p.uniqueStrings(module.Dependencies)
}

// extractModuleName extracts module name from tags
func (p *AdvancedTagParser) extractModuleName(tags []Tag) string {
	for _, tag := range tags {
		if tag.Type == TagTypeModule {
			if tag.Value != "" {
				return tag.Value
			}
		}
	}
	return ""
}

// isProjectImport checks if an import is from the current project
func (p *AdvancedTagParser) isProjectImport(importPath, projectPath string) bool {
	// This is a simplified check - could be enhanced
	return strings.Contains(importPath, "/internal/") ||
		strings.Contains(importPath, "/cmd/") ||
		strings.HasPrefix(importPath, "github.com/pmaojo/kthulu-go/backend/")
}

// extractModuleFromImport extracts module name from import path
func (p *AdvancedTagParser) extractModuleFromImport(importPath string) string {
	// Extract module name from path like "github.com/pmaojo/kthulu-go/backend/internal/modules/user"
	parts := strings.Split(importPath, "/")
	if len(parts) >= 4 && parts[len(parts)-2] == "modules" {
		return parts[len(parts)-1]
	}
	return ""
}

// uniqueStrings removes duplicate strings from slice
func (p *AdvancedTagParser) uniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// shouldIgnoreFile checks if file should be ignored based on patterns
func (p *AdvancedTagParser) shouldIgnoreFile(filePath string) bool {
	for _, pattern := range p.config.IgnorePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}
		if strings.Contains(filePath, pattern) {
			return true
		}
	}

	// Default ignores
	defaultIgnores := []string{"vendor/", ".git/", "_test.go", "testdata/"}
	for _, ignore := range defaultIgnores {
		if strings.Contains(filePath, ignore) {
			return true
		}
	}

	return false
}

// generateCacheKey generates a cache key for a project
func (p *AdvancedTagParser) generateCacheKey(projectPath string) string {
	hash := sha256.Sum256([]byte(projectPath))
	return fmt.Sprintf("project:%x", hash)
}

// generateFileCacheKey generates a cache key for a file
func (p *AdvancedTagParser) generateFileCacheKey(filePath string) string {
	// Include file modification time in hash for cache invalidation
	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Sprintf("file:%x", sha256.Sum256([]byte(filePath)))
	}

	content := filePath + ":" + stat.ModTime().String()
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("file:%x", hash)
}
