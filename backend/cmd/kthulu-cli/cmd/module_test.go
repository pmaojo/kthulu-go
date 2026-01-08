package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldModule(t *testing.T) {
	dir := t.TempDir()
	if err := scaffoldModule(dir, "foo"); err != nil {
		t.Fatalf("scaffoldModule failed: %v", err)
	}
	path := filepath.Join(dir, "backend", "internal", "modules", "foo.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading module: %v", err)
	}
	if !strings.Contains(string(data), "FooModule") {
		t.Fatalf("module file missing name: %s", string(data))
	}
}
