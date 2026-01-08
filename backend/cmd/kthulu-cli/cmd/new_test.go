//go:build integration
// +build integration

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestScaffoldProjectGoVersion(t *testing.T) {
	dir := t.TempDir()
	project := filepath.Join(dir, "proj")
	if err := scaffoldProject(project, nil, false); err != nil {
		t.Fatalf("scaffoldProject failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(project, "go.mod"))
	if err != nil {
		t.Fatalf("reading go.mod: %v", err)
	}
	want := "go " + goVersion
	if !strings.Contains(string(data), want) {
		t.Fatalf("go.mod does not contain %q: %s", want, string(data))
	}
}

func TestScaffoldProjectCreatesFiles(t *testing.T) {
	tmp := t.TempDir()
	base := filepath.Join(tmp, "myapp")
	if err := scaffoldProject(base, nil, false); err != nil {
		t.Fatalf("scaffoldProject failed: %v", err)
	}

	files := []string{
		"go.mod",
		"docker-compose.yml",
		"Makefile",
		"README.md",
		filepath.Join("backend", "cmd", "service", "main.go"),
	}
	for _, f := range files {
		p := filepath.Join(base, f)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s to exist: %v", p, err)
		}
	}

	goMod, err := os.ReadFile(filepath.Join(base, "go.mod"))
	if err != nil {
		t.Fatalf("reading go.mod: %v", err)
	}
	if !strings.Contains(string(goMod), "module myapp") {
		t.Fatalf("go.mod missing module name: %s", string(goMod))
	}

	readme, err := os.ReadFile(filepath.Join(base, "README.md"))
	if err != nil {
		t.Fatalf("reading README.md: %v", err)
	}
	if !strings.Contains(string(readme), "# myapp") {
		t.Fatalf("README.md missing module name: %s", string(readme))
	}
}

func TestScaffoldProjectSkipExisting(t *testing.T) {
	tmp := t.TempDir()
	base := filepath.Join(tmp, "myapp")
	if err := scaffoldProject(base, nil, false); err != nil {
		t.Fatalf("initial scaffoldProject failed: %v", err)
	}

	prePath := filepath.Join(base, "go.mod")
	original := []byte("original content\n")
	if err := os.WriteFile(prePath, original, 0o644); err != nil {
		t.Fatalf("writing preexisting file: %v", err)
	}

	if err := scaffoldProject(base, nil, true); err != nil {
		t.Fatalf("scaffoldProject with skipExisting failed: %v", err)
	}

	b, err := os.ReadFile(prePath)
	if err != nil {
		t.Fatalf("reading preexisting file: %v", err)
	}
	if string(b) != string(original) {
		t.Fatalf("expected preexisting file to remain unchanged, got: %s", string(b))
	}
}

func TestCopyDirFSCopiesFiles(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "file.txt"), []byte("hi"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "nested.txt"), []byte("nested"), 0640); err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	dst := filepath.Join(tmp, "out")
	if err := copyDirFS(os.DirFS(src), ".", dst, false); err != nil {
		t.Fatalf("copyDirFS: %v", err)
	}
	if b, err := os.ReadFile(filepath.Join(dst, "file.txt")); err != nil || string(b) != "hi" {
		t.Fatalf("file not copied: %v %s", err, string(b))
	}
	if _, err := os.Stat(filepath.Join(dst, "sub", "nested.txt")); err != nil {
		t.Fatalf("nested file missing: %v", err)
	}
}

func TestCopyFileFSSkipExisting(t *testing.T) {
	fsys := fstest.MapFS{
		"file.txt": {Data: []byte("new"), Mode: 0o600},
	}
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "file.txt")
	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatalf("prewrite: %v", err)
	}
	if err := copyFileFS(fsys, "file.txt", dst, true); err != nil {
		t.Fatalf("copyFileFS: %v", err)
	}
	b, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(b) != "old" {
		t.Fatalf("expected old content, got %s", string(b))
	}
}

func TestCopyModuleCopiesAssets(t *testing.T) {
	tmp := t.TempDir()
	if err := copyModule("access", tmp, false); err != nil {
		t.Fatalf("copyModule access: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "backend", "internal", "modules", "access.go")); err != nil {
		t.Fatalf("backend module not copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "frontend", "src", "modules", "access", "index.ts")); err != nil {
		t.Fatalf("frontend module not copied: %v", err)
	}
}
