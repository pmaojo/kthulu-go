package depanalysis

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyze(t *testing.T) {
	t.Run("valid layering", func(t *testing.T) {
		root := filepath.Join("testdata", "valid")
		if err := Analyze(root); err != nil {
			t.Fatalf("Analyze() error: %v", err)
		}
	})

	t.Run("layer violation", func(t *testing.T) {
		root := filepath.Join("testdata", "violates")
		err := Analyze(root)
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "layer violation") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("import cycle", func(t *testing.T) {
		root := filepath.Join("testdata", "cycle")
		err := Analyze(root)
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "import cycle") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
