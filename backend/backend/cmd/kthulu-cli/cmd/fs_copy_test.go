package cmd

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"
)

func TestCopyFileFSPreservesMode(t *testing.T) {
	fsys := fstest.MapFS{
		"file.txt": {Data: []byte("hello"), Mode: 0o641},
	}
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "file.txt")
	if err := copyFileFS(fsys, "file.txt", dst, false); err != nil {
		t.Fatalf("copyFileFS: %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dst: %v", err)
	}
	if info.Mode() != 0o641 {
		t.Fatalf("mode = %v, want 0641", info.Mode())
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("content = %q, want %q", string(data), "hello")
	}
}

type dirReadlinkFS struct {
	fs.FS
	root string
}

func (d dirReadlinkFS) Readlink(name string) (string, error) {
	return os.Readlink(filepath.Join(d.root, name))
}

func TestCopyFileFSSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated permissions on Windows")
	}
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "target.txt"), []byte("t"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.Symlink("target.txt", filepath.Join(srcDir, "link")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	dst := filepath.Join(t.TempDir(), "link")
	fsys := dirReadlinkFS{FS: os.DirFS(srcDir), root: srcDir}
	if err := copyFileFS(fsys, "link", dst, false); err != nil {
		t.Fatalf("copyFileFS: %v", err)
	}
	if got, err := os.Readlink(dst); err != nil || got != "target.txt" {
		t.Fatalf("symlink copied incorrectly: %q %v", got, err)
	}
}

func TestCopyDirFSSkipSpecialDirs(t *testing.T) {
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, ".git", "config"), []byte("cfg"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "node_modules", "pkg"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "node_modules", "pkg", "index.js"), []byte("pkg"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "dir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "dir", "file.txt"), []byte("hi"), 0600); err != nil {
		t.Fatal(err)
	}

	dst := t.TempDir()
	if err := copyDirFS(os.DirFS(src), ".", dst, false); err != nil {
		t.Fatalf("copyDirFS: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, ".git")); !os.IsNotExist(err) {
		t.Fatalf(".git should not be copied")
	}
	if _, err := os.Stat(filepath.Join(dst, "node_modules")); !os.IsNotExist(err) {
		t.Fatalf("node_modules should not be copied")
	}
	if _, err := os.Stat(filepath.Join(dst, "dir", "file.txt")); err != nil {
		t.Fatalf("expected file copied: %v", err)
	}
}

func TestCopyDirFSSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated permissions on Windows")
	}
	src := t.TempDir()
	if err := os.Mkdir(filepath.Join(src, "dir"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "dir", "file.txt"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.Symlink("dir/file.txt", filepath.Join(src, "link")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	dst := t.TempDir()
	fsys := dirReadlinkFS{FS: os.DirFS(src), root: src}
	if err := copyDirFS(fsys, ".", dst, false); err != nil {
		t.Fatalf("copyDirFS: %v", err)
	}
	if got, err := os.Readlink(filepath.Join(dst, "link")); err != nil || got != "dir/file.txt" {
		t.Fatalf("symlink copied incorrectly: %q %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(dst, "dir", "file.txt")); err != nil {
		t.Fatalf("file not copied: %v", err)
	}
}
