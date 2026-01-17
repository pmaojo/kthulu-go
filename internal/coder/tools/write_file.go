package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// WriteFileTool writes content to a file
var WriteFileTool = &Tool{
	Name: "write_file",
	Description: `Write content to a file on the filesystem.

Use this tool to:
- Create new files
- Overwrite existing files completely
- Save generated code or configuration

The tool will:
- Create parent directories if they don't exist
- Show a diff if the file already exists
- Require approval before writing

For small edits to existing files, consider describing the changes instead of rewriting the entire file.`,
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	},
	NeedsApproval: true,
	Execute:       executeWriteFile,
}

func executeWriteFile(ctx context.Context, args map[string]any) (*Result, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return &Result{
			Success: false,
			Error:   "path is required",
		}, nil
	}

	content, ok := args["content"].(string)
	if !ok {
		return &Result{
			Success: false,
			Error:   "content is required",
		}, nil
	}

	// Resolve path
	if !filepath.IsAbs(path) {
		cwd, _ := os.Getwd()
		path = filepath.Join(cwd, path)
	}

	// Check if file exists and get old content for diff
	var oldContent string
	isNew := true
	if data, err := os.ReadFile(path); err == nil {
		oldContent = string(data)
		isNew = false
	}

	// Create parent directories
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("failed to create directory: %v", err),
		}, nil
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}

	// Build output
	var output, display string
	if isNew {
		lineCount := len(strings.Split(content, "\n"))
		output = fmt.Sprintf("Created file %s (%d lines)", path, lineCount)
		display = formatNewFileOutput(path, content)
	} else {
		output = fmt.Sprintf("Updated file %s", path)
		display = formatDiffOutput(path, oldContent, content)
	}

	return &Result{
		Success: true,
		Output:  output,
		Display: display,
	}, nil
}

func formatNewFileOutput(path, content string) string {
	var sb strings.Builder

	filename := filepath.Base(path)
	lines := strings.Split(content, "\n")
	lineCount := len(lines)

	sb.WriteString(fmt.Sprintf("╭─ [NEW] %s (%d lines) ", filename, lineCount))
	sb.WriteString(strings.Repeat("─", 30))
	sb.WriteString("╮\n")

	// Show first 10 lines
	maxLines := 10
	for i, line := range lines {
		if i >= maxLines {
			sb.WriteString(fmt.Sprintf("│ ... (+%d more lines)\n", lineCount-maxLines))
			break
		}
		sb.WriteString(fmt.Sprintf("│ %s\n", line))
	}

	sb.WriteString("╰" + strings.Repeat("─", 50) + " ✓ Created ─╯\n")

	return sb.String()
}

func formatDiffOutput(path, oldContent, newContent string) string {
	var sb strings.Builder

	filename := filepath.Base(path)

	sb.WriteString(fmt.Sprintf("╭─ [MODIFIED] %s ", filename))
	sb.WriteString(strings.Repeat("─", 35))
	sb.WriteString("╮\n")

	// Compute diff
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldContent, newContent, false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	// Count changes
	additions := 0
	deletions := 0
	for _, diff := range diffs {
		lines := len(strings.Split(diff.Text, "\n")) - 1
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			additions += lines
		case diffmatchpatch.DiffDelete:
			deletions += lines
		}
	}

	// Show unified diff style
	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		for i, line := range lines {
			// Skip empty trailing line
			if i == len(lines)-1 && line == "" {
				continue
			}
			switch diff.Type {
			case diffmatchpatch.DiffInsert:
				sb.WriteString(fmt.Sprintf("│ + %s\n", line))
			case diffmatchpatch.DiffDelete:
				sb.WriteString(fmt.Sprintf("│ - %s\n", line))
			default:
				// Show context around changes (skip long unchanged sections)
				if len(line) > 0 {
					sb.WriteString(fmt.Sprintf("│   %s\n", line))
				}
			}
		}
	}

	sb.WriteString(fmt.Sprintf("╰─ +%d/-%d ", additions, deletions))
	sb.WriteString(strings.Repeat("─", 40))
	sb.WriteString(" ✓ Updated ─╯\n")

	return sb.String()
}
