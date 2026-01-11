package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pmaojo/kthulu-go/cmd/kthulu/templates"
)

type FrontendScaffolder struct {
	projectPath string
	fs          FileSystem
}

func NewFrontendScaffolder(projectPath string) *FrontendScaffolder {
	return &FrontendScaffolder{
		projectPath: projectPath,
		fs:          RealFileSystem{},
	}
}

// EnsureFrontend checks if frontend exists, if not scaffolds it
func (s *FrontendScaffolder) EnsureFrontend() error {
	frontendPath := filepath.Join(s.projectPath, "frontend")
	if _, err := os.Stat(frontendPath); !os.IsNotExist(err) {
		// Frontend exists, check if we need to check for App.tsx?
		// For now assume it's good.
		return nil
	}

	fmt.Println("Frontend not found. Scaffolding Base Admin UI...")

	// Copy from templates/frontend/base -> frontend/
	baseDir := "frontend/base"

	return s.copyRecursive(baseDir, frontendPath)
}

func (s *FrontendScaffolder) copyRecursive(srcDir, destDir string) error {
	entries, err := templates.FS.ReadDir(srcDir)
	if err != nil {
		return err
	}

	if err := s.fs.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := srcDir + "/" + entry.Name()
		destPath := filepath.Join(destDir, strings.TrimSuffix(entry.Name(), ".tmpl"))

		if entry.IsDir() {
			if err := s.copyRecursive(srcPath, destPath); err != nil {
				return err
			}
		} else {
			content, err := templates.FS.ReadFile(srcPath)
			if err != nil {
				return err
			}

			// If it's a template, we might want to execute it (e.g. package.json name)
			// For now, simple copy or generic execution
			if strings.HasSuffix(entry.Name(), ".tmpl") {
				// Execute as template if needed, or just write.
				// Since base templates don't use much dynamic data (maybe project name?),
				// I'll execute with empty map for now or a generic struct.
				// But some files might have {{ }} for React/JSX code!
				// WARNING: Go templates conflict with JSX {{ }}.
				// My templates used {{ }} for Go, but JSX uses { }.
				// I should be careful.
				// If the file is .tsx.tmpl, I need to assume it might contain template tags.
				// But looking at my written files, I didn't use any {{ .Var }} in the base templates yet, except maybe package.json?
				// Actually I wrote them as raw strings in previous steps.
				// I should probably just WriteFile raw content if no template logic is needed.
				// BUT `templates.go` logic usually strips .tmpl.

				// HACK: Just write content directly for now to avoid JSX/Template conflicts
				// unless I explicitly know it needs interpolation.
				// My `App.tsx` has `/* KTHULU:ROUTES */` which is NOT a template tag.
				// So I can just write it.

				if err := s.fs.WriteFile(destPath, content, 0644); err != nil {
					return err
				}
			} else {
				if err := s.fs.WriteFile(destPath, content, 0644); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
