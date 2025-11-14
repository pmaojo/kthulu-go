package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildAndWritePlan(t *testing.T) {
	constructs := []Construct{
		{ID: "root", Path: "module", Priority: 100},
		{ID: "file1", Path: "module/a.go", Priority: 80},
		{ID: "file1-deco", Path: "module/a.go", Priority: 70},
		{ID: "subroot", Path: "module/sub", Priority: 60},
		{ID: "subfile", Path: "module/sub/b.go", Priority: 50},
	}

	p := Build(constructs)
	dir := t.TempDir()
	if err := Write(p, dir); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, ".kthulu", "plan.json"))
	if err != nil {
		t.Fatalf("read plan: %v", err)
	}
	want, err := os.ReadFile(filepath.Join("testdata", "plan_snapshot.json"))
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("plan JSON mismatch\n---got---\n%s\n---want---\n%s", got, want)
	}
}
