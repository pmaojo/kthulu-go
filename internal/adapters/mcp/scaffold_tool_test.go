package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateScaffoldArgsRejectsFieldlessModules(t *testing.T) {
	cases := []struct {
		name string
		args ScaffoldProjectArgs
		want string
	}{
		{"no name", ScaffoldProjectArgs{}, "name is required"},
		{"no modules", ScaffoldProjectArgs{Name: "x"}, "modules is required"},
		{"empty fields", ScaffoldProjectArgs{Name: "x", Modules: []ScaffoldModule{{Name: "tournament"}}}, "has no fields"},
		{"missing type", ScaffoldProjectArgs{Name: "x", Modules: []ScaffoldModule{{Name: "t", Fields: []string{"title"}}}}, "missing a type"},
	}
	for _, tc := range cases {
		err := validateScaffoldArgs(tc.args)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s: expected error containing %q, got %v", tc.name, tc.want, err)
		}
	}

	ok := ScaffoldProjectArgs{Name: "x", Modules: []ScaffoldModule{{Name: "t", Fields: []string{"title:string:required"}}}}
	if err := validateScaffoldArgs(ok); err != nil {
		t.Fatalf("expected valid args, got %v", err)
	}
}

func TestWriteScaffoldPlan(t *testing.T) {
	dir := t.TempDir()
	path, err := writeScaffoldPlan(dir, ScaffoldProjectArgs{
		Name:     "tourney",
		Features: []string{"auth", "queues"},
		Modules: []ScaffoldModule{
			{Name: "team", Fields: []string{"name:string:required", "wins:int"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("plan written outside working dir: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{"name: tourney", "database: sqlite", "team:", "name:string:required", "queues"} {
		if !strings.Contains(content, want) {
			t.Fatalf("plan missing %q:\n%s", want, content)
		}
	}
}
