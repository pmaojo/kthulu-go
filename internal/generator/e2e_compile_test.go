package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/internal/resolver"
	"github.com/stretchr/testify/require"
)

// TestE2E_GeneratedProjectCompiles generates a full project on disk and
// compiles it with the Go toolchain. It downloads module dependencies, so it
// only runs when KTHULU_E2E=1 is set (e.g. in CI with network access).
//
//	KTHULU_E2E=1 go test ./internal/generator/ -run TestE2E -timeout 20m
func TestE2E_GeneratedProjectCompiles(t *testing.T) {
	if os.Getenv("KTHULU_E2E") != "1" {
		t.Skip("set KTHULU_E2E=1 to run the end-to-end compile test (requires network)")
	}

	root := t.TempDir()
	outputPath := filepath.Join(root, "shopapp")

	gen := NewTemplateGenerator(resolver.NewDependencyResolver(&parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}))

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "shopapp",
		ProjectModule: "github.com/e2e/shopapp",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "templ",
		OutputPath:    outputPath,
		Features:      []string{"auth", "user", "product", "payments", "queues"},
		ModuleFields: map[string][]string{
			"product": {"name:string:required,min=2", "price:int:required,min=1", "in_stock:bool", "contact_email:string:email"},
		},
		CustomValues: map[string]string{"module_path": "github.com/e2e/shopapp"},
	})
	require.NoError(t, err)
	require.NoError(t, gen.WriteProject(structure))

	run := func(name string, args ...string) {
		t.Helper()
		cmd := exec.Command(name, args...)
		cmd.Dir = outputPath
		cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s %v failed:\n%s", name, args, out)
	}

	run("go", "run", "github.com/a-h/templ/cmd/templ@v0.3.977", "generate", "./...")
	run("go", "mod", "tidy")
	run("go", "build", "./...")
	run("go", "vet", "./...")
	// Exercises the generated queue runtime tests (processing, retries,
	// dead-lettering) inside the generated project.
	run("go", "test", "./internal/infrastructure/queue/...")
}
