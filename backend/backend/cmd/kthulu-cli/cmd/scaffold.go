package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/templates"
)

const goVersion = "1.24"

// writeTemplate renders the named template from the embedded template set into dst.
// When skipExisting is true and dst already exists, the file is left untouched.
func writeTemplate(templatePath, dst string, data any, skipExisting bool) error {
	if skipExisting {
		if _, err := os.Stat(dst); err == nil {
			return nil
		}
	}

	tplData, err := templates.Templates.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read template %s: %w", templatePath, err)
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(tplData))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", templatePath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template %s: %w", templatePath, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	return os.WriteFile(dst, buf.Bytes(), 0o644)
}

// copyTemplateFile copies a static file from the embedded template set to dst.
func copyTemplateFile(templatePath, dst string, skipExisting bool) error {
	if skipExisting {
		if _, err := os.Stat(dst); err == nil {
			return nil
		}
	}

	contents, err := templates.Templates.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read template file %s: %w", templatePath, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	return os.WriteFile(dst, contents, 0o644)
}

// copyTemplateTree copies a directory tree from the embedded template set, rendering
// any files that end with .tmpl and preserving the directory structure.
func copyTemplateTree(fsys fs.FS, src, dst string, data map[string]any, skipExisting bool) error {
	if _, err := fs.Stat(fsys, src); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	return fs.WalkDir(fsys, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := dst
		if rel != "." {
			target = filepath.Join(dst, rel)
		}

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(target, info.Mode())
		}

		if strings.HasSuffix(d.Name(), ".tmpl") {
			return writeTemplate(path, strings.TrimSuffix(target, ".tmpl"), data, skipExisting)
		}

		return copyTemplateFile(path, target, skipExisting)
	})
}

// scaffoldProject creates a new project skeleton under base using the embedded templates.
func scaffoldProject(base string, modules []string, skipExisting bool) error {
	if strings.TrimSpace(base) == "" {
		return fmt.Errorf("invalid project path")
	}

	if err := os.MkdirAll(base, 0o755); err != nil {
		return err
	}

	projectName := filepath.Base(base)
	templateData := map[string]any{
		"ModuleName": projectName,
		"GoVersion":  goVersion,
	}

	templatedFiles := map[string]string{
		"go.mod.tmpl":             "go.mod",
		"README.md.tmpl":          "README.md",
		"Makefile.tmpl":           "Makefile",
		"docker-compose.yml.tmpl": "docker-compose.yml",
	}

	for src, dst := range templatedFiles {
		if err := writeTemplate(src, filepath.Join(base, dst), templateData, skipExisting); err != nil {
			return err
		}
	}

	directories := []struct {
		src string
		dst string
	}{
		{"backend", filepath.Join(base, "backend")},
		{"frontend", filepath.Join(base, "frontend")},
		{"internal", filepath.Join(base, "internal")},
		{"migrations", filepath.Join(base, "migrations")},
		{"openapi", filepath.Join(base, "openapi")},
	}

	for _, dir := range directories {
		if err := copyTemplateTree(templates.Templates, dir.src, dir.dst, templateData, skipExisting); err != nil {
			return err
		}
	}

	extraFiles := []struct {
		src string
		dst string
	}{
		{"deployment.yaml", filepath.Join(base, "deployment.yaml")},
		{"deployment-info.txt", filepath.Join(base, "deployment-info.txt")},
	}

	for _, file := range extraFiles {
		if err := copyTemplateFile(file.src, file.dst, skipExisting); err != nil {
			return err
		}
	}

	for _, module := range modules {
		if err := copyModule(module, base, skipExisting); err != nil {
			return err
		}
	}

	return nil
}

// copyModule copies backend and frontend assets for the named module into base.
func copyModule(name, base string, skipExisting bool) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("invalid module name")
	}

	backendFile := filepath.Join("internal", name+".go")
	if _, err := fs.Stat(templates.Templates, backendFile); err != nil {
		return fmt.Errorf("backend template for module %s: %w", name, err)
	}

	backendDst := filepath.Join(base, "backend", "internal", "modules", name+".go")
	if err := copyTemplateFile(backendFile, backendDst, skipExisting); err != nil {
		return err
	}

	if info, err := fs.Stat(templates.Templates, filepath.Join("internal", name)); err == nil && info.IsDir() {
		dst := filepath.Join(base, "backend", "internal", "modules", name)
		if err := copyTemplateTree(templates.Templates, filepath.Join("internal", name), dst, nil, skipExisting); err != nil {
			return err
		}
	}

	if info, err := fs.Stat(templates.Templates, filepath.Join("frontend", "src", "modules", name)); err == nil && info.IsDir() {
		dst := filepath.Join(base, "frontend", "src", "modules", name)
		if err := copyTemplateTree(templates.Templates, filepath.Join("frontend", "src", "modules", name), dst, nil, skipExisting); err != nil {
			return err
		}
	}

	return nil
}

// scaffoldModule generates a backend module file from the template using the provided name.
func scaffoldModule(base, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("nombre de módulo inválido")
	}

	data := map[string]any{
		"ModuleName":   name,
		"ModuleExport": exportName(name),
	}

	dst := filepath.Join(base, "backend", "internal", "modules", name+".go")
	return writeTemplate("module.go.tmpl", dst, data, false)
}
