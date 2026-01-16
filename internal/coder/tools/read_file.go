package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileTool reads file contents
var ReadFileTool = &Tool{
	Name: "read_file",
	Description: `Read the contents of a file from the filesystem.

Use this tool to:
- Understand existing code before making changes
- Check configuration files
- Review documentation

The file path should be absolute or relative to the working directory.
Returns the file contents with line numbers for easy reference.`,
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to read",
			},
			"start_line": map[string]interface{}{
				"type":        "integer",
				"description": "Start line number (1-indexed, optional)",
			},
			"end_line": map[string]interface{}{
				"type":        "integer",
				"description": "End line number (1-indexed, optional)",
			},
		},
		"required": []string{"path"},
	},
	NeedsApproval: false,
	Execute:       executeReadFile,
}

func executeReadFile(ctx context.Context, args map[string]any) (*Result, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return &Result{
			Success: false,
			Error:   "path is required",
		}, nil
	}

	// Resolve path
	if !filepath.IsAbs(path) {
		cwd, _ := os.Getwd()
		path = filepath.Join(cwd, path)
	}

	// Check file exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return &Result{
			Success: false,
			Output:  fmt.Sprintf("File not found: %s", path),
			Error:   "file not found",
		}, nil
	}
	if info.IsDir() {
		return &Result{
			Success: false,
			Output:  fmt.Sprintf("%s is a directory, not a file", path),
			Error:   "path is a directory",
		}, nil
	}

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return &Result{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	lines := strings.Split(string(content), "\n")
	totalLines := len(lines)

	// Parse line range
	startLine := 1
	endLine := totalLines
	if sl, ok := args["start_line"].(float64); ok && sl > 0 {
		startLine = int(sl)
	}
	if el, ok := args["end_line"].(float64); ok && el > 0 {
		endLine = int(el)
	}

	// Clamp to valid range
	if startLine < 1 {
		startLine = 1
	}
	if endLine > totalLines {
		endLine = totalLines
	}
	if startLine > endLine {
		startLine = endLine
	}

	// Extract lines and add line numbers
	var sb strings.Builder
	for i := startLine - 1; i < endLine; i++ {
		sb.WriteString(fmt.Sprintf("%4d │ %s\n", i+1, lines[i]))
	}

	output := sb.String()

	// Create display output
	display := formatFileOutput(path, output, startLine, endLine, totalLines)

	return &Result{
		Success: true,
		Output:  output,
		Display: display,
	}, nil
}

func formatFileOutput(path, content string, startLine, endLine, totalLines int) string {
	var sb strings.Builder

	filename := filepath.Base(path)
	header := fmt.Sprintf("╭─ %s (lines %d-%d of %d) ", filename, startLine, endLine, totalLines)
	sb.WriteString(header)
	sb.WriteString(strings.Repeat("─", max(0, 60-len(header))))
	sb.WriteString("╮\n")

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if line != "" {
			sb.WriteString(fmt.Sprintf("│%s\n", line))
		}
	}

	sb.WriteString("╰" + strings.Repeat("─", 60) + "╯\n")

	return sb.String()
}
