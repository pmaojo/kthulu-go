package builder

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// helper to write plan struct to file
func writePlan(t *testing.T, p Plan, path string) {
	t.Helper()
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
}

func TestGenerateProducesCompiledAndContracts(t *testing.T) {
	outDir := filepath.Join("testdata", "generated", t.Name())
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(filepath.Join("testdata", "generated")) })

	plan := Plan{
		Replacements: []Replacement{
			{
				Interface:      "github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/builder/testdata/interfaces.Service",
				Implementation: "github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/builder/testdata/overrides.MockService",
				Constructor:    "github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/builder/testdata/overrides.NewMockService",
			},
		},
		Decorations: []string{
			"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/builder/testdata/decorator.DecorateService",
		},
		Groups: map[string][]string{
			"hooks": []string{"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/builder/testdata/hooks.NewHook"},
		},
	}
	planPath := filepath.Join(outDir, "plan.json")
	writePlan(t, plan, planPath)

	if err := Generate(planPath, outDir); err != nil {
		t.Fatalf("generate: %v", err)
	}

	// ensure compiled.go has key fx constructs
	compiledBytes, err := os.ReadFile(filepath.Join(outDir, "compiled.go"))
	if err != nil {
		t.Fatalf("read compiled: %v", err)
	}
	compiled := string(compiledBytes)
	if !strings.Contains(compiled, "fx.Replace") || !strings.Contains(compiled, "fx.Decorate") || !strings.Contains(compiled, "group:\"hooks\"") {
		t.Fatalf("compiled.go missing expected constructs: %s", compiled)
	}

	// run go test on generated package to ensure compilation
	cmd := exec.Command("go", "test", "./testdata/generated/"+t.Name())
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test failed: %v\n%s", err, out)
	}
}

func TestGenerateContractFails(t *testing.T) {
	outDir := filepath.Join("testdata", "generated", t.Name())
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(filepath.Join("testdata", "generated")) })

	plan := Plan{
		Replacements: []Replacement{
			{
				Interface:      "github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/builder/testdata/interfaces.Service",
				Implementation: "github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/builder/testdata/badoverride.BadService",
				Constructor:    "github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/builder/testdata/badoverride.NewBadService",
			},
		},
	}
	planPath := filepath.Join(outDir, "plan.json")
	writePlan(t, plan, planPath)

	if err := Generate(planPath, outDir); err != nil {
		t.Fatalf("generate: %v", err)
	}

	cmd := exec.Command("go", "test", "./testdata/generated/"+t.Name())
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected go test to fail\n%s", out)
	}
}
