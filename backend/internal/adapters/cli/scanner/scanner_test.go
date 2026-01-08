package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestScan_GoAndYAML(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "overrides", "a.go"), `package test

// @kthulu:shadow module:mod/sym symbol:Func priority:10
func Func() {}
`)

	writeFile(t, filepath.Join(root, "extends", "b.yaml"), `kthulu:
  wrap:
    module: mod/sym
    symbol: Other
    priority: 5
`)

	anns, err := Scan(root)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(anns) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(anns))
	}
}

func TestScan_Duplicate(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "overrides", "a.go"), `package test
// @kthulu:shadow module:mod symbol:X
`)
	writeFile(t, filepath.Join(root, "extends", "b.go"), `package test
// @kthulu:shadow module:mod symbol:X
`)

	if _, err := Scan(root); err == nil {
		t.Fatalf("expected duplicate error")
	}
}

func TestScan_InvalidAnnotations(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "overrides", "a.go"), `package test
// @kthulu:shadow symbol:X
`)

	if _, err := Scan(root); err == nil {
		t.Fatalf("expected error for invalid comment annotation")
	}

	root2 := t.TempDir()
	writeFile(t, filepath.Join(root2, "extends", "b.yaml"), `kthulu:
  shadow:
    module: mod
`)

	if _, err := Scan(root2); err == nil {
		t.Fatalf("expected error for invalid yaml annotation")
	}
}
