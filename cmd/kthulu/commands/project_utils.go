package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getProjectModule(dir string) (string, error) {
	goModPath := filepath.Join(dir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", nil
}

func runGoModTidy(projectPath string) error {
	fmt.Println("\n🧹 Running go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	cmd.Env = append(os.Environ(), "GOWORK=off")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runTemplGenerate(projectPath string) error {
	fmt.Println("\n✨ Generating GTH (Templ) components...")
	// We use 'go run' so we don't require the user to have the templ binary installed globally.
	// We use @v0.2.793 to match our internal version and ensure compatibility.
	cmd := exec.Command("go", "run", "github.com/a-h/templ/cmd/templ@v0.2.793", "generate", "./...")
	cmd.Dir = projectPath
	cmd.Env = append(os.Environ(), "GOWORK=off")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
