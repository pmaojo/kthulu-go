package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBashTool_Execute(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantOK  bool
		wantOut string
	}{
		{
			name:    "simple echo",
			args:    map[string]any{"command": "echo hello"},
			wantOK:  true,
			wantOut: "hello",
		},
		{
			name:   "empty command",
			args:   map[string]any{"command": ""},
			wantOK: false,
		},
		{
			name:   "missing command",
			args:   map[string]any{},
			wantOK: false,
		},
		{
			name:    "exit code preserved",
			args:    map[string]any{"command": "exit 1"},
			wantOK:  false,
			wantOut: "",
		},
		{
			name:    "stderr captured",
			args:    map[string]any{"command": "echo error >&2"},
			wantOK:  true,
			wantOut: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BashTool.Execute(context.Background(), tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Success != tt.wantOK {
				t.Errorf("Success = %v, want %v", result.Success, tt.wantOK)
			}
			if tt.wantOut != "" && !strings.Contains(result.Output, tt.wantOut) {
				t.Errorf("Output = %q, want to contain %q", result.Output, tt.wantOut)
			}
		})
	}
}

func TestReadFileTool_Execute(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line 1\nline 2\nline 3\nline 4\nline 5"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		args    map[string]any
		wantOK  bool
		wantOut string
	}{
		{
			name:    "read entire file",
			args:    map[string]any{"path": testFile},
			wantOK:  true,
			wantOut: "line 1",
		},
		{
			name:    "read with line range",
			args:    map[string]any{"path": testFile, "start_line": 2.0, "end_line": 3.0},
			wantOK:  true,
			wantOut: "line 2",
		},
		{
			name:   "file not found",
			args:   map[string]any{"path": "/nonexistent/file.txt"},
			wantOK: false,
		},
		{
			name:   "missing path",
			args:   map[string]any{},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ReadFileTool.Execute(context.Background(), tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Success != tt.wantOK {
				t.Errorf("Success = %v, want %v", result.Success, tt.wantOK)
			}
			if tt.wantOut != "" && !strings.Contains(result.Output, tt.wantOut) {
				t.Errorf("Output = %q, want to contain %q", result.Output, tt.wantOut)
			}
		})
	}
}

func TestWriteFileTool_Execute(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		args    map[string]any
		setup   func() // Optional setup
		wantOK  bool
		verify  func(t *testing.T) // Optional verification
	}{
		{
			name: "create new file",
			args: map[string]any{
				"path":    filepath.Join(tmpDir, "new.txt"),
				"content": "hello world",
			},
			wantOK: true,
			verify: func(t *testing.T) {
				data, err := os.ReadFile(filepath.Join(tmpDir, "new.txt"))
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != "hello world" {
					t.Errorf("file content = %q, want %q", string(data), "hello world")
				}
			},
		},
		{
			name: "overwrite existing file",
			args: map[string]any{
				"path":    filepath.Join(tmpDir, "existing.txt"),
				"content": "new content",
			},
			setup: func() {
				os.WriteFile(filepath.Join(tmpDir, "existing.txt"), []byte("old content"), 0644)
			},
			wantOK: true,
			verify: func(t *testing.T) {
				data, _ := os.ReadFile(filepath.Join(tmpDir, "existing.txt"))
				if string(data) != "new content" {
					t.Errorf("file content = %q, want %q", string(data), "new content")
				}
			},
		},
		{
			name: "create with nested dirs",
			args: map[string]any{
				"path":    filepath.Join(tmpDir, "a", "b", "c", "file.txt"),
				"content": "nested",
			},
			wantOK: true,
		},
		{
			name: "missing path",
			args: map[string]any{
				"content": "hello",
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result, err := WriteFileTool.Execute(context.Background(), tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Success != tt.wantOK {
				t.Errorf("Success = %v, want %v", result.Success, tt.wantOK)
			}
			if tt.verify != nil {
				tt.verify(t)
			}
		})
	}
}

func TestThinkTool_Execute(t *testing.T) {
	tests := []struct {
		name   string
		args   map[string]any
		wantOK bool
	}{
		{
			name:   "valid thought",
			args:   map[string]any{"thought": "I need to refactor this..."},
			wantOK: true,
		},
		{
			name:   "empty thought",
			args:   map[string]any{"thought": ""},
			wantOK: false,
		},
		{
			name:   "missing thought",
			args:   map[string]any{},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ThinkTool.Execute(context.Background(), tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Success != tt.wantOK {
				t.Errorf("Success = %v, want %v", result.Success, tt.wantOK)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	t.Run("all default tools registered", func(t *testing.T) {
		expectedTools := []string{"bash", "read_file", "write_file", "grep", "think", "kthulu"}
		for _, name := range expectedTools {
			if _, ok := r.Get(name); !ok {
				t.Errorf("expected tool %q to be registered", name)
			}
		}
	})

	t.Run("get returns correct tool", func(t *testing.T) {
		tool, ok := r.Get("bash")
		if !ok {
			t.Fatal("expected bash tool to exist")
		}
		if tool.Name != "bash" {
			t.Errorf("tool name = %q, want %q", tool.Name, "bash")
		}
	})

	t.Run("get returns false for unknown tool", func(t *testing.T) {
		_, ok := r.Get("nonexistent")
		if ok {
			t.Error("expected Get to return false for unknown tool")
		}
	})

	t.Run("ToOpenAIFormat", func(t *testing.T) {
		format := r.ToOpenAIFormat()
		if len(format) == 0 {
			t.Error("expected non-empty OpenAI format")
		}
		for _, f := range format {
			if f["type"] != "function" {
				t.Errorf("type = %v, want 'function'", f["type"])
			}
		}
	})
}
