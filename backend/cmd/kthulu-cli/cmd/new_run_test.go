package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRunGoTestsAllowsPartialCoverage(t *testing.T) {
	dir := t.TempDir()
	writeModuleFile(t, dir, "go.mod", "module example.com/testproj\n\ngo 1.24\n")
	writeModuleFile(t, dir, "main.go", "package main\n\nfunc add(a, b int) int { return a + b }\nfunc sub(a, b int) int { return a - b }\n\nfunc main() {}\n")
	writeModuleFile(t, dir, "main_test.go", "package main\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n    if add(2, 3) != 5 {\n        t.Fatalf(\"unexpected sum\")\n    }\n}\n")

	if err := runGoTests(dir); err != nil {
		t.Fatalf("runGoTests should pass with partial coverage: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "coverage.out")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("coverage.out should be removed after running go tests, got: %v", err)
	}
}

func TestRunGoTestsFailsWhenTestsFail(t *testing.T) {
	dir := t.TempDir()
	writeModuleFile(t, dir, "go.mod", "module example.com/testproj\n\ngo 1.24\n")
	writeModuleFile(t, dir, "main.go", "package main\n\nfunc alwaysFalse() bool { return false }\n")
	writeModuleFile(t, dir, "main_test.go", "package main\n\nimport \"testing\"\n\nfunc TestFail(t *testing.T) {\n    if alwaysFalse() {\n        t.Fatal(\"unexpected truthy\")\n    } else {\n        t.Fatal(\"expected failure for coverage enforcement\")\n    }\n}\n")

	if err := runGoTests(dir); err == nil {
		t.Fatalf("runGoTests should fail when go test fails")
	}
}

func writeModuleFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
}
