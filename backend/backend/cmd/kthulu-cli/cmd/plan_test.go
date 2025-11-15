package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	planpkg "github.com/pmaojo/kthulu-go/backend/internal/cli/plan"
	graphpkg "github.com/pmaojo/kthulu-go/backend/internal/graph"
)

func TestPlanCommand(t *testing.T) {
	rootDir := t.TempDir()
	overrides := filepath.Join(rootDir, "overrides")
	if err := os.MkdirAll(overrides, 0o755); err != nil {
		t.Fatalf("mkdir overrides: %v", err)
	}
	if err := os.WriteFile(filepath.Join(overrides, "a.go"), []byte("package test\n// @kthulu:shadow module:mod symbol:X priority:10\n"), 0o644); err != nil {
		t.Fatalf("write override: %v", err)
	}

	root := &cobra.Command{Use: "root"}
	root.AddCommand(newPlanCmd())
	root.SetArgs([]string{"plan", rootDir})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute plan: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(rootDir, ".kthulu", "plan.json"))
	if err != nil {
		t.Fatalf("read plan: %v", err)
	}
	var p planpkg.Plan
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("unmarshal plan: %v", err)
	}
	if len(p.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(p.Nodes))
	}
	n := p.Nodes[0]
	if n.Construct.Path != filepath.Join("mod", "X") {
		t.Fatalf("unexpected path: %s", n.Construct.Path)
	}
	if n.Construct.Priority != 10 {
		t.Fatalf("unexpected priority: %d", n.Construct.Priority)
	}
	if n.Action != planpkg.Replace {
		t.Fatalf("unexpected action: %s", n.Action)
	}
}

func TestPlanCommandGraph(t *testing.T) {
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("JWT_REFRESH_SECRET", "s")
	t.Setenv("JWT_REFRESH_TOKEN_TTL", "1h")
	os.Remove("/tmp/kthulu.graph.json")

	root := &cobra.Command{Use: "root"}
	root.AddCommand(newPlanCmd())
	root.SetArgs([]string{"plan", "--graph", "--format=json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute plan with graph: %v", err)
	}
	data, err := os.ReadFile("/tmp/kthulu.graph.json")
	if err != nil {
		t.Fatalf("read graph: %v", err)
	}
	var g graphpkg.Graph
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("unmarshal graph: %v", err)
	}
}

func TestPlanCommandValidate(t *testing.T) {
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("JWT_REFRESH_SECRET", "s")
	t.Setenv("JWT_REFRESH_TOKEN_TTL", "1h")

	root := &cobra.Command{Use: "root"}
	root.AddCommand(newPlanCmd())
	root.SetArgs([]string{"plan", "--validate"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute plan validate: %v", err)
	}
}
func TestPlanCommandDepAnalysisFailures(t *testing.T) {
	cases := []struct {
		name string
		base string
	}{
		{
			name: "layer violation",
			base: filepath.Join("..", "..", "..", "internal", "cli", "depanalysis", "testdata", "violates"),
		},
		{
			name: "import cycle",
			base: filepath.Join("..", "..", "..", "internal", "cli", "depanalysis", "testdata", "cycle"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// ensure leftover plan does not exist
			os.RemoveAll(filepath.Join(tc.base, ".kthulu"))
			root := &cobra.Command{Use: "root"}
			root.AddCommand(newPlanCmd())
			root.SetArgs([]string{"plan", tc.base})
			if err := root.Execute(); err == nil {
				t.Fatalf("expected error")
			}
			if _, err := os.Stat(filepath.Join(tc.base, ".kthulu", "plan.json")); !os.IsNotExist(err) {
				t.Fatalf("plan.json should not be generated")
			}
		})
	}
}
