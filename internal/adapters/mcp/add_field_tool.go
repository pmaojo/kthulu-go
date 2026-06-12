package mcp

import (
	"context"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// AddFieldArgs for the add_field tool.
type AddFieldArgs struct {
	Module string `json:"module" jsonschema:"required,description=Module name (e.g. 'order')."`
	Field  string `json:"field"  jsonschema:"required,description=Field definition: name:type[:rules]. E.g. 'shipped_at:time' or 'discount_cents:int:min=0' or 'category:belongs_to:category'."`
}

// validFieldTypes is the set of accepted type tokens.
var validFieldTypes = map[string]bool{
	"string":     true,
	"int":        true,
	"float":      true,
	"bool":       true,
	"time":       true,
	"belongs_to": true,
}

// snakeToCamel converts a snake_case identifier to CamelCase.
func snakeToCamel(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' })
	var b strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	if b.Len() == 0 {
		return "Field"
	}
	return b.String()
}

// camelToSnake converts a CamelCase identifier to snake_case.
func camelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// parseAddFieldSpec splits a field definition string into (name, type, rules).
// rules is the raw remainder after name:type (may be empty).
func parseAddFieldSpec(field string) (name, typ, rules string, err error) {
	parts := strings.SplitN(field, ":", 3)
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("field must be name:type[:rules], got %q", field)
	}
	name = strings.TrimSpace(parts[0])
	typ = strings.TrimSpace(parts[1])
	if len(parts) == 3 {
		rules = strings.TrimSpace(parts[2])
	}
	if name == "" {
		return "", "", "", fmt.Errorf("field name must not be empty")
	}
	if !validFieldTypes[typ] {
		return "", "", "", fmt.Errorf("unknown field type %q; valid types: string int float bool time belongs_to", typ)
	}
	return name, typ, rules, nil
}

// findModelFile searches for internal/<module>/model.go case-insensitively.
func findModelFile(workingDir, module string) (string, error) {
	internalDir := filepath.Join(workingDir, "internal")
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return "", fmt.Errorf("cannot read internal/: %w", err)
	}

	moduleLower := strings.ToLower(module)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.ToLower(e.Name()) == moduleLower {
			candidate := filepath.Join(internalDir, e.Name(), "model.go")
			if _, statErr := os.Stat(candidate); statErr == nil {
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("model.go not found for module %q (searched internal/%s/model.go)", module, module)
}

// findMainStruct finds the primary exported struct in the file — preferring the one
// that embeds gorm.Model or whose name matches the module name.
func findMainStruct(fset *token.FileSet, file *ast.File, module string) *ast.StructType {
	modulePascal := snakeToCamel(module)
	modulePascalLower := strings.ToLower(modulePascal)

	var best *ast.StructType
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			// Must be exported.
			if len(typeSpec.Name.Name) == 0 || !unicode.IsUpper(rune(typeSpec.Name.Name[0])) {
				continue
			}

			// Exact name match — take it immediately.
			if strings.ToLower(typeSpec.Name.Name) == modulePascalLower {
				return st
			}

			// Contains gorm.Model — remember as a fallback.
			if best == nil && hasGormModel(st) {
				best = st
			}
		}
	}
	// If we found a gorm.Model struct but no name match, return it.
	if best != nil {
		return best
	}
	// Last resort: first exported struct in the file.
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			if len(typeSpec.Name.Name) > 0 && unicode.IsUpper(rune(typeSpec.Name.Name[0])) {
				return st
			}
		}
	}
	return nil
}

// findMainStructName returns the name of the primary exported struct.
func findMainStructName(fset *token.FileSet, file *ast.File, module string) string {
	modulePascal := snakeToCamel(module)
	modulePascalLower := strings.ToLower(modulePascal)

	var bestName string
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			_, ok = typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			if len(typeSpec.Name.Name) == 0 || !unicode.IsUpper(rune(typeSpec.Name.Name[0])) {
				continue
			}
			if strings.ToLower(typeSpec.Name.Name) == modulePascalLower {
				return typeSpec.Name.Name
			}
			if bestName == "" {
				bestName = typeSpec.Name.Name
			}
		}
	}
	return bestName
}

// hasGormModel returns true when a struct directly embeds gorm.Model.
func hasGormModel(st *ast.StructType) bool {
	if st.Fields == nil {
		return false
	}
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			// Embedded field — check selector or ident.
			switch t := f.Type.(type) {
			case *ast.SelectorExpr:
				if t.Sel.Name == "Model" {
					return true
				}
			case *ast.Ident:
				if t.Name == "Model" {
					return true
				}
			}
		}
	}
	return false
}

// buildValidateTag constructs the validate struct tag value from rules.
// Returns "" when there are no rules.
func buildValidateTag(rules string) string {
	if rules == "" {
		return ""
	}
	// Replace | separators in oneof= with pipes — they arrive as-is from the DSL.
	return rules
}

// buildFieldLines returns one or more lines to append inside the struct.
// For belongs_to it returns two lines (FK + relation).
func buildFieldLines(fieldName, fieldType, rules, module string) ([]string, error) {
	// Use the raw fieldName for snake (input is already expected to be snake_case).
	snakeRaw := fieldName

	goName := snakeToCamel(fieldName)
	validateTag := buildValidateTag(rules)

	switch fieldType {
	case "belongs_to":
		// rules holds the related module name for belongs_to.
		relModule := rules
		if relModule == "" {
			relModule = fieldName
		}
		relPascal := snakeToCamel(relModule)

		// FK field
		fkSnake := snakeRaw + "_id"
		fkGoName := goName + "ID"
		fkLine := fmt.Sprintf(
			"\t%s uint `gorm:\"column:%s;index\" json:\"%s\"`",
			fkGoName, fkSnake, fkSnake,
		)

		// Relation field
		relLine := fmt.Sprintf(
			"\t%s *%s `gorm:\"foreignKey:%s\" json:\"%s,omitempty\"`",
			relPascal, relPascal, fkGoName, snakeRaw,
		)

		return []string{fkLine, relLine}, nil

	case "time":
		tag := fmt.Sprintf("`gorm:\"column:%s\" json:\"%s\"`", snakeRaw, snakeRaw)
		if validateTag != "" {
			tag = fmt.Sprintf("`gorm:\"column:%s\" json:\"%s\" validate:\"%s\"`", snakeRaw, snakeRaw, validateTag)
		}
		line := fmt.Sprintf("\t%s *time.Time %s", goName, tag)
		return []string{line}, nil

	case "bool":
		tag := fmt.Sprintf("`gorm:\"column:%s;default:false\" json:\"%s\"`", snakeRaw, snakeRaw)
		if validateTag != "" {
			tag = fmt.Sprintf("`gorm:\"column:%s;default:false\" json:\"%s\" validate:\"%s\"`", snakeRaw, snakeRaw, validateTag)
		}
		line := fmt.Sprintf("\t%s bool %s", goName, tag)
		return []string{line}, nil

	case "string":
		gormParts := fmt.Sprintf("column:%s", snakeRaw)
		if strings.Contains(validateTag, "required") {
			gormParts += ";not null"
		}
		tag := fmt.Sprintf("`gorm:\"%s\" json:\"%s\"`", gormParts, snakeRaw)
		if validateTag != "" {
			tag = fmt.Sprintf("`gorm:\"%s\" json:\"%s\" validate:\"%s\"`", gormParts, snakeRaw, validateTag)
		}
		line := fmt.Sprintf("\t%s string %s", goName, tag)
		return []string{line}, nil

	case "int":
		gormParts := fmt.Sprintf("column:%s", snakeRaw)
		if strings.Contains(validateTag, "required") {
			gormParts += ";not null"
		}
		tag := fmt.Sprintf("`gorm:\"%s\" json:\"%s\"`", gormParts, snakeRaw)
		if validateTag != "" {
			tag = fmt.Sprintf("`gorm:\"%s\" json:\"%s\" validate:\"%s\"`", gormParts, snakeRaw, validateTag)
		}
		line := fmt.Sprintf("\t%s int %s", goName, tag)
		return []string{line}, nil

	case "float":
		gormParts := fmt.Sprintf("column:%s", snakeRaw)
		if strings.Contains(validateTag, "required") {
			gormParts += ";not null"
		}
		tag := fmt.Sprintf("`gorm:\"%s\" json:\"%s\"`", gormParts, snakeRaw)
		if validateTag != "" {
			tag = fmt.Sprintf("`gorm:\"%s\" json:\"%s\" validate:\"%s\"`", gormParts, snakeRaw, validateTag)
		}
		line := fmt.Sprintf("\t%s float64 %s", goName, tag)
		return []string{line}, nil

	default:
		return nil, fmt.Errorf("unsupported field type %q", fieldType)
	}
}

// insertFieldsIntoSource inserts new field lines before the closing brace of
// the target struct. It uses text manipulation on the original source so that
// comments and formatting are preserved, then re-formats with go/format.
func insertFieldsIntoSource(src []byte, fset *token.FileSet, st *ast.StructType, newLines []string) ([]byte, error) {
	if st.Fields == nil {
		return nil, fmt.Errorf("struct has no field list")
	}

	// The closing brace position.
	closingPos := fset.Position(st.Fields.Closing)
	closingOffset := closingPos.Offset

	insertion := "\n" + strings.Join(newLines, "\n") + "\n"

	modified := make([]byte, 0, len(src)+len(insertion))
	modified = append(modified, src[:closingOffset]...)
	modified = append(modified, []byte(insertion)...)
	modified = append(modified, src[closingOffset:]...)

	// Re-format.
	formatted, err := format.Source(modified)
	if err != nil {
		return nil, fmt.Errorf("go/format failed after insert: %w", err)
	}
	return formatted, nil
}

// sqlPreviewForField produces a rough SQL ALTER TABLE preview for the new fields.
func sqlPreviewForField(tableName, fieldName, fieldType, rules string) string {
	snake := fieldName
	switch fieldType {
	case "belongs_to":
		fkSnake := snake + "_id"
		relPascal := snakeToCamel(rules)
		if rules == "" {
			relPascal = snakeToCamel(snake)
		}
		return fmt.Sprintf(
			"  ALTER TABLE %s ADD COLUMN %s INTEGER;\n  -- (relation field %s is not a DB column)",
			tableName, fkSnake, relPascal,
		)
	case "time":
		return fmt.Sprintf("  ALTER TABLE %s ADD COLUMN %s TIMESTAMP;", tableName, snake)
	case "bool":
		return fmt.Sprintf("  ALTER TABLE %s ADD COLUMN %s BOOLEAN DEFAULT FALSE;", tableName, snake)
	case "string":
		return fmt.Sprintf("  ALTER TABLE %s ADD COLUMN %s TEXT;", tableName, snake)
	case "int":
		return fmt.Sprintf("  ALTER TABLE %s ADD COLUMN %s INTEGER;", tableName, snake)
	case "float":
		return fmt.Sprintf("  ALTER TABLE %s ADD COLUMN %s REAL;", tableName, snake)
	}
	return ""
}

// pluralise is a simple English pluraliser for table name derivation.
// It mirrors what GORM does by default (lowercase + plural).
func pluralise(s string) string {
	s = strings.ToLower(s)
	switch {
	case strings.HasSuffix(s, "y") && len(s) > 1:
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") ||
		strings.HasSuffix(s, "sh") || strings.HasSuffix(s, "ch"):
		return s + "es"
	}
	return s + "s"
}

// addFieldTool returns the add_field RegisteredTool.
func addFieldTool(executor CommandExecutor, workingDir string) RegisteredTool {
	return RegisteredTool{
		Name: "add_field",
		Description: "Add a new field to an existing module's model struct and generate the corresponding migration. " +
			"Safer than editing model.go manually — handles GORM tags, JSON tags, and validation tags automatically.",
		Handler: func(ctx context.Context, args AddFieldArgs) (*mcp_golang.ToolResponse, error) {
			dir := resolveWorkdir(workingDir)

			// 1. Parse and validate.
			module := strings.TrimSpace(args.Module)
			if module == "" {
				return nil, fmt.Errorf("module is required")
			}
			fieldRaw := strings.TrimSpace(args.Field)
			if fieldRaw == "" {
				return nil, fmt.Errorf("field is required")
			}

			fieldName, fieldType, rules, err := parseAddFieldSpec(fieldRaw)
			if err != nil {
				return nil, err
			}

			// 2. Find model.go.
			modelPath, err := findModelFile(dir, module)
			if err != nil {
				return nil, err
			}

			// 3. Read file.
			src, err := os.ReadFile(modelPath)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", modelPath, err)
			}
			originalSrc := make([]byte, len(src))
			copy(originalSrc, src)

			// 4. Parse AST.
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, modelPath, src, parser.ParseComments)
			if err != nil {
				return nil, fmt.Errorf("parse %s: %w", modelPath, err)
			}

			// 5. Find the target struct.
			st := findMainStruct(fset, astFile, module)
			if st == nil {
				return nil, fmt.Errorf("no exported struct found in %s", modelPath)
			}
			structName := findMainStructName(fset, astFile, module)

			// 6. Build field lines.
			newLines, err := buildFieldLines(fieldName, fieldType, rules, module)
			if err != nil {
				return nil, err
			}

			// 7. Insert into source.
			modified, err := insertFieldsIntoSource(src, fset, st, newLines)
			if err != nil {
				return nil, err
			}

			// 8. Write modified file.
			if err := os.WriteFile(modelPath, modified, 0o644); err != nil {
				return nil, fmt.Errorf("write %s: %w", modelPath, err)
			}

			// 9. Verify compilation — build the module package.
			moduleDir := filepath.Dir(modelPath)
			buildResult, buildErr := executor.Run(ctx, dir, []string{"build", "./internal/..." })
			// If build fails, restore original and return error.
			if buildErr != nil {
				_ = os.WriteFile(modelPath, originalSrc, 0o644)
				buildOut := strings.TrimSpace(buildResult.Stdout)
				if strings.TrimSpace(buildResult.Stderr) != "" {
					buildOut += "\n" + strings.TrimSpace(buildResult.Stderr)
				}
				return nil, fmt.Errorf("compilation failed after inserting field (model.go restored):\n%s", buildOut)
			}
			_ = moduleDir // used implicitly via executor.Run path above

			// 10. Run migrate diff --dry-run for preview.
			migrateResult, _ := executor.Run(ctx, dir, []string{"migrate", "diff", "--dry-run"})
			migrateOut := strings.TrimSpace(migrateResult.Stdout)
			if strings.TrimSpace(migrateResult.Stderr) != "" {
				migrateOut += "\n" + strings.TrimSpace(migrateResult.Stderr)
			}
			if migrateOut == "" {
				// Fall back to a static preview when migrate diff is unavailable.
				tableName := pluralise(snakeToCamel(module))
				tableName = camelToSnake(tableName)
				migrateOut = sqlPreviewForField(tableName, fieldName, fieldType, rules)
			}

			// 11. Build summary.
			rel, _ := filepath.Rel(dir, modelPath)

			// Describe what was added.
			addedLines := strings.Join(newLines, "\n")
			// Strip leading tab for display.
			var displayLines []string
			for _, l := range newLines {
				displayLines = append(displayLines, strings.TrimPrefix(l, "\t"))
			}

			summary := fmt.Sprintf(
				"Added field '%s' to %s model (%s)\n\nNew field:\n  %s\n\nMigration preview:\n%s\n\nNext: run migrate_preview with save=true (or kthulu migrate diff) to write the migration file.",
				fieldName,
				structName,
				rel,
				strings.Join(displayLines, "\n  "),
				migrateOut,
			)

			_ = addedLines // used above

			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(summary)), nil
		},
	}
}

func init() {
	RegisterPlugin(func(executor CommandExecutor, workingDir string) []RegisteredTool {
		return []RegisteredTool{addFieldTool(executor, workingDir)}
	})
}
