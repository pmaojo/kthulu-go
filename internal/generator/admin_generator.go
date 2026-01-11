package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/templates"
)

// AdminGenerator handles the generation of admin UI components
type AdminGenerator struct {
	projectPath string
	fs          FileSystem // Abstraction for testing if needed
}

type FileSystem interface {
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
}

type RealFileSystem struct{}

func (fs RealFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (fs RealFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// NewAdminGenerator creates a new AdminGenerator
func NewAdminGenerator(projectPath string) *AdminGenerator {
	return &AdminGenerator{
		projectPath: projectPath,
		fs:          RealFileSystem{},
	}
}

// GenerateAdminModule generates the admin UI for a given entity definition
func (g *AdminGenerator) GenerateAdminModule(def UIEntityDefinition) error {
	// Target Directory: frontend/src/modules/<module>/presentation/admin/<Entity>
	targetDir := filepath.Join(g.projectPath, "frontend", "src", "modules", def.Module, "presentation", "admin", def.Name)
	if err := g.fs.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	// Template Files (name in embed fs -> target path)
	tmplFiles := map[string]string{
		"table.tsx.tmpl": filepath.Join(targetDir, def.Name+"Table.tsx"),
		"form.tsx.tmpl":  filepath.Join(targetDir, def.Name+"Form.tsx"),
		"page.tsx.tmpl":  filepath.Join(targetDir, "index.tsx"), // Entry point for the entity page
		"api.ts.tmpl":    filepath.Join(targetDir, "api.ts"),
	}

	// Function map for templates
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	for tmplName, targetPath := range tmplFiles {
		// Read from Embedded FS
		content, err := templates.FS.ReadFile("frontend/admin/" + tmplName)
		if err != nil {
			return fmt.Errorf("failed to read embedded template %s: %w", tmplName, err)
		}

		t, err := template.New(tmplName).Funcs(funcMap).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", tmplName, err)
		}

		var buf bytes.Buffer
		if err := t.Execute(&buf, def); err != nil {
			return fmt.Errorf("failed to execute template %s: %w", tmplName, err)
		}

		if err := g.fs.WriteFile(targetPath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetPath, err)
		}
		fmt.Printf("Generated %s\n", targetPath)
	}

	return nil
}
