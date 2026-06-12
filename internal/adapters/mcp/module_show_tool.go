package mcp

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// ModuleShowArgs for module_show tool.
type ModuleShowArgs struct {
	Name string `json:"name" jsonschema:"required,description=Module name (e.g. 'order', 'user'). Case-insensitive."`
}

// moduleField holds parsed field info from a model struct.
type moduleField struct {
	goName  string
	goType  string
	gormTag string
	jsonTag string
	valTag  string
}

// moduleShowTool returns the module_show RegisteredTool bound to workingDir.
func moduleShowTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name: "module_show",
		Description: "Show complete introspection of a project module: fields/types/validations from the model struct, " +
			"service methods, handler functions, related modules. " +
			"Like 'artisan model:show' for kthulu projects.",
		Handler: func(ctx context.Context, args ModuleShowArgs) (*mcp_golang.ToolResponse, error) {
			if strings.TrimSpace(args.Name) == "" {
				return nil, fmt.Errorf("name is required")
			}

			dir := resolveWorkdir(workingDir)
			modDir, modName, err := findModuleDir(dir, args.Name)
			if err != nil {
				return nil, err
			}

			relModDir, relErr := filepath.Rel(dir, modDir)
			if relErr != nil {
				relModDir = modDir
			}

			report, err := buildModuleReport(dir, modDir, relModDir, modName)
			if err != nil {
				return nil, err
			}

			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(report)), nil
		},
	}
}

// findModuleDir locates the internal/<name> directory, case-insensitively.
// Returns the absolute dir path, the canonical module name (dir basename), and
// an error listing available modules when not found.
func findModuleDir(projectDir, name string) (string, string, error) {
	lower := strings.ToLower(name)

	// Try direct match first.
	candidate := filepath.Join(projectDir, "internal", lower)
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate, lower, nil
	}

	// Scan internal/ for a case-insensitive match.
	internalDir := filepath.Join(projectDir, "internal")
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return "", "", fmt.Errorf("cannot read internal/: %w", err)
	}

	var available []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		available = append(available, e.Name())
		if strings.EqualFold(e.Name(), name) {
			full := filepath.Join(internalDir, e.Name())
			return full, e.Name(), nil
		}
	}

	return "", "", fmt.Errorf("module %q not found in internal/\nAvailable modules: %s",
		name, strings.Join(available, ", "))
}

// buildModuleReport assembles the full textual report for a module.
func buildModuleReport(projectDir, modDir, relModDir, modName string) (string, error) {
	var sb strings.Builder

	// Capitalise first letter for display.
	displayName := strings.ToUpper(modName[:1]) + modName[1:]
	sb.WriteString(fmt.Sprintf("📦 Module: %s (%s/)\n", displayName, relModDir))

	// --- model.go ---
	modelPath := filepath.Join(modDir, "model.go")
	if _, err := os.Stat(modelPath); err == nil {
		fields, err := parseModelFields(modelPath)
		if err == nil && len(fields) > 0 {
			sb.WriteString("\nFIELDS (from model.go):\n")
			for _, f := range fields {
				// Column name: prefer gorm column tag, else json tag, else snake_case name.
				col := columnName(f)
				line := fmt.Sprintf("  %-16s%-12s", col, f.goType)
				extras := []string{}
				if f.gormTag != "" {
					extras = append(extras, fmt.Sprintf("gorm:%q", f.gormTag))
				}
				if f.valTag != "" {
					extras = append(extras, fmt.Sprintf("validate:%q", f.valTag))
				}
				if f.jsonTag != "" && f.jsonTag != col {
					extras = append(extras, fmt.Sprintf("→ json: %s", f.jsonTag))
				}
				if len(extras) > 0 {
					line += strings.Join(extras, "  ")
				}
				sb.WriteString(line + "\n")
			}
		}
	}

	// --- service.go ---
	servicePath := filepath.Join(modDir, "service.go")
	if _, err := os.Stat(servicePath); err == nil {
		methods, err := parseExportedFunctions(servicePath)
		if err == nil && len(methods) > 0 {
			sb.WriteString("\nSERVICE METHODS (service.go):\n")
			for _, m := range methods {
				sb.WriteString(fmt.Sprintf("  %s\n", m))
			}
		}
	}

	// --- handler.go ---
	handlerPath := filepath.Join(modDir, "handler.go")
	if _, err := os.Stat(handlerPath); err == nil {
		handlers, err := parseExportedFunctionNames(handlerPath)
		if err == nil && len(handlers) > 0 {
			sb.WriteString("\nHANDLERS (handler.go):\n")
			sb.WriteString(fmt.Sprintf("  %s\n", strings.Join(handlers, ", ")))
		}
	}

	// --- migrations/ ---
	migCount := countMigrationFiles(projectDir, modName)
	sb.WriteString("\nMIGRATIONS: ")
	if migCount == 0 {
		sb.WriteString("no SQL files reference this module\n")
	} else {
		sb.WriteString(fmt.Sprintf("%d SQL file(s) reference this module\n", migCount))
	}

	return sb.String(), nil
}

// parseModelFields parses the first exported struct in a model.go and returns
// its fields with GORM / validate / json tag info.
func parseModelFields(path string) ([]moduleField, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			// Only exported structs.
			if !ts.Name.IsExported() {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			return extractStructFields(st), nil
		}
	}
	return nil, nil
}

// extractStructFields walks every field of a struct and pulls tag info.
func extractStructFields(st *ast.StructType) []moduleField {
	var fields []moduleField
	for _, field := range st.Fields.List {
		typStr := types.ExprString(field.Type)

		var rawTag string
		if field.Tag != nil {
			rawTag = strings.Trim(field.Tag.Value, "`")
		}

		gormTag := extractTagValue(rawTag, "gorm")
		jsonTag := extractTagValue(rawTag, "json")
		valTag := extractTagValue(rawTag, "validate")

		// Strip omitempty from json tag.
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonTag = jsonTag[:idx]
		}

		if len(field.Names) == 0 {
			// Embedded field – use type name.
			fields = append(fields, moduleField{
				goName:  typStr,
				goType:  typStr,
				gormTag: gormTag,
				jsonTag: jsonTag,
				valTag:  valTag,
			})
			continue
		}
		for _, name := range field.Names {
			fields = append(fields, moduleField{
				goName:  name.Name,
				goType:  typStr,
				gormTag: gormTag,
				jsonTag: jsonTag,
				valTag:  valTag,
			})
		}
	}
	return fields
}

// extractTagValue pulls the value for a specific tag key out of a raw struct
// tag string (the content between backticks, without the backticks themselves).
func extractTagValue(raw, key string) string {
	// Look for key:"..." or key:`...`
	prefix := key + `:"`
	idx := strings.Index(raw, prefix)
	if idx == -1 {
		return ""
	}
	rest := raw[idx+len(prefix):]
	end := strings.Index(rest, `"`)
	if end == -1 {
		return ""
	}
	return rest[:end]
}

// columnName returns the display column name for a field: gorm column tag >
// json tag > snake_case of the Go field name.
func columnName(f moduleField) string {
	if col := extractTagPart(f.gormTag, "column"); col != "" {
		return col
	}
	if f.jsonTag != "" && f.jsonTag != "-" {
		return f.jsonTag
	}
	return toSnakeCase(f.goName)
}

// extractTagPart picks out a sub-key from a GORM tag value like
// "column:total_cents;not null".
func extractTagPart(gormVal, key string) string {
	for _, part := range strings.Split(gormVal, ";") {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 && strings.EqualFold(kv[0], key) {
			return kv[1]
		}
	}
	return ""
}

// toSnakeCase converts CamelCase to snake_case (best-effort for display).
func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + 32) // toLower
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// parseExportedFunctions returns signature strings for exported top-level
// functions and methods in a Go file (no body, just name + params + results).
func parseExportedFunctions(path string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var sigs []string
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || !fd.Name.IsExported() {
			continue
		}
		sig := fd.Name.Name + strings.TrimPrefix(types.ExprString(fd.Type), "func")
		sigs = append(sigs, sig)
	}
	return sigs, nil
}

// parseExportedFunctionNames returns just the names of exported top-level
// functions (and methods) in a Go file.
func parseExportedFunctionNames(path string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || !fd.Name.IsExported() {
			continue
		}
		names = append(names, fd.Name.Name)
	}
	return names, nil
}

// countMigrationFiles counts .sql files in migrations/ whose filename contains
// the module name (case-insensitive).
func countMigrationFiles(projectDir, modName string) int {
	migrDir := filepath.Join(projectDir, "migrations")
	entries, err := os.ReadDir(migrDir)
	if err != nil {
		return 0
	}
	lower := strings.ToLower(modName)
	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		if strings.Contains(strings.ToLower(e.Name()), lower) {
			count++
		}
	}
	return count
}

func init() {
	RegisterPlugin(func(_ CommandExecutor, workingDir string) []RegisteredTool {
		return []RegisteredTool{moduleShowTool(workingDir)}
	})
}
