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

	// Copy from templates/scaffold/frontend/base -> frontend/
	baseDir := "scaffold/frontend/base"

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
			// Just write content directly. If template processing is needed later, add specific checks.
			if err := s.fs.WriteFile(destPath, content, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
