package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetWorkdir() {
	sessionWorkdir.mu.Lock()
	sessionWorkdir.path = ""
	sessionWorkdir.mu.Unlock()
}

func TestSessionWorkdirOverride(t *testing.T) {
	t.Cleanup(resetWorkdir)
	resetWorkdir()

	base := t.TempDir()
	if got := resolveWorkdir(base); got != base {
		t.Fatalf("expected configured dir %s, got %s", base, got)
	}

	project := filepath.Join(base, "myapp")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}

	// Relative paths resolve against the current working directory.
	dir, err := setSessionWorkdir(base, "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if dir != project {
		t.Fatalf("expected %s, got %s", project, dir)
	}
	if got := resolveWorkdir(base); got != project {
		t.Fatalf("override not applied: %s", got)
	}

	// Nonexistent directories are rejected and the override is kept.
	if _, err := setSessionWorkdir(base, "missing-dir"); err == nil {
		t.Fatal("expected error for missing directory")
	}
	if got := resolveWorkdir(base); got != project {
		t.Fatalf("override lost after failed set: %s", got)
	}

	// Empty path resets to the configured default.
	if _, err := setSessionWorkdir(base, ""); err != nil {
		t.Fatal(err)
	}
	if got := resolveWorkdir(base); got != base {
		t.Fatalf("expected reset to %s, got %s", base, got)
	}
}

func TestDescribeWorkdirDetectsProject(t *testing.T) {
	t.Cleanup(resetWorkdir)
	resetWorkdir()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "kthulu-plan.yaml"), []byte("name: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	desc := describeWorkdir(dir)
	if !strings.Contains(desc, "kthulu project") {
		t.Fatalf("expected kthulu project detection, got: %s", desc)
	}
	if !strings.Contains(desc, "kthulu-plan.yaml") {
		t.Fatalf("expected file listing, got: %s", desc)
	}
}
