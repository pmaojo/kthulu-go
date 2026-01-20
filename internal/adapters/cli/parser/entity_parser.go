package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
)

// EntityParser handles parsing of Go structs to extract entity definitions
type EntityParser struct {
	projectPath string
}

// NewEntityParser creates a new EntityParser
func NewEntityParser(projectPath string) *EntityParser {
	return &EntityParser{
		projectPath: projectPath,
	}
}

// ParseFile parses a single Go file and returns discovered entities
func (p *EntityParser) ParseFile(filePath string) ([]EntityDefinition, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	var entities []EntityDefinition

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for type definitions
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		// We only care about structs
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Check if it's an entity (heuristic: has ID or specific tags)
		if p.isEntity(structType, typeSpec.Name.Name) {
			entity := p.parseEntity(typeSpec.Name.Name, structType)
			entities = append(entities, entity)
		}

		return false // Don't traverse inside the struct
	})

	return entities, nil
}

// isEntity determines if a struct is an entity worth generating an admin UI for
func (p *EntityParser) isEntity(structType *ast.StructType, name string) bool {
	// Heuristic 1: Has an "ID" field
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if name.Name == "ID" {
				return true
			}
		}
	}

	// TODO: Check for doc comments / tags once we have access to the full comment map
	// For now, implicit check based on ID is a good start for the MVP
	return false
}

func (p *EntityParser) parseEntity(name string, structType *ast.StructType) EntityDefinition {
	entity := EntityDefinition{
		Name: name,
	}

	for _, field := range structType.Fields.List {
		// Handle embedded fields or multiple names
		if len(field.Names) == 0 {
			// Embedded field (e.g. generic.Model)
			// For now, skip or handle basic cases if needed.
			// Real-world entities in Kthulu seem to have explicit ID fields.
			continue
		}

		for _, fieldName := range field.Names {
			def := FieldDefinition{
				Name: fieldName.Name,
				Type: p.extractType(field.Type),
			}

			if field.Tag != nil {
				// Strip backticks
				tagValue := strings.Trim(field.Tag.Value, "`")
				tags := reflect.StructTag(tagValue)

				// Parse JSON tag
				if jsonTag := tags.Get("json"); jsonTag != "" {
					parts := strings.Split(jsonTag, ",")
					def.JSONName = parts[0]
				}

				// Parse Validate tag
				if validateTag := tags.Get("validate"); validateTag != "" {
					def.ValidationRules = p.parseValidation(validateTag)
				}
			}

			// Detect relationships
			if p.isSliceOfStructs(field.Type) {
				def.IsRelation = true
				def.RelationType = "HasMany" // Simplified
			} else if p.isStructPointer(field.Type) {
				def.IsRelation = true
				def.RelationType = "BelongsTo" // Simplified
			}

			entity.Fields = append(entity.Fields, def)
		}
	}

	return entity
}

func (p *EntityParser) extractType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + p.extractType(t.X)
	case *ast.SelectorExpr:
		return p.extractType(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + p.extractType(t.Elt)
	case *ast.MapType:
		return "map[" + p.extractType(t.Key) + "]" + p.extractType(t.Value)
	default:
		return "unknown"
	}
}

func (p *EntityParser) isSliceOfStructs(expr ast.Expr) bool {
	arrayType, ok := expr.(*ast.ArrayType)
	if !ok {
		return false
	}
	// Check if element is a struct (Identifier that resolves to struct, or just Identifier)
	// Without full type resolution, we assume non-basic types are potential structs
	typeName := p.extractType(arrayType.Elt)
	return !isBasicType(typeName)
}

func (p *EntityParser) isStructPointer(expr ast.Expr) bool {
	starExpr, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	typeName := p.extractType(starExpr.X)
	return !isBasicType(typeName)
}

var basicTypes = map[string]bool{
	"string": true, "int": true, "int64": true, "uint": true, "uint64": true,
	"float64": true, "bool": true, "byte": true, "time.Time": true, // treating time as basic for this check
}

func isBasicType(t string) bool {
	return basicTypes[t]
}

func (p *EntityParser) parseValidation(tag string) []string {
	return strings.Split(tag, ",")
}
