package flags

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadHeaderConfigFrom(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "headers.yml")
	content := "X-Test-Flag: test_flag\nX-Another: another_flag\n"
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := LoadHeaderConfigFrom(file)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg["X-Test-Flag"] != "test_flag" {
		t.Errorf("expected test_flag, got %q", cfg["X-Test-Flag"])
	}
	if cfg["X-Another"] != "another_flag" {
		t.Errorf("expected another_flag, got %q", cfg["X-Another"])
	}
}

func TestLoadHeaderConfigMissing(t *testing.T) {
	cfg, err := LoadHeaderConfigFrom("nonexistent.yml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg) != 0 {
		t.Fatalf("expected empty config, got %v", cfg)
	}
}
