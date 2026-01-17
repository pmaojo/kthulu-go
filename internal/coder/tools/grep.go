package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GrepTool searches for patterns in files using ripgrep
var GrepTool = &Tool{
	Name: "grep",
	Description: `Search for patterns in files using ripgrep.

Use this tool to:
- Find files containing specific patterns
- Search for function definitions
- Locate usages of variables or types
- Find TODO comments or specific strings

Features:
- Fast regex search across the entire codebase
- Respects .gitignore by default
- Returns matching file paths sorted by modification time`,
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "The regex pattern to search for",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory to search in (default: current directory)",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "File pattern to include (e.g., '*.go', '*.ts')",
			},
		},
		"required": []string{"pattern"},
	},
	NeedsApproval: false,
	Execute:       executeGrep,
}

func executeGrep(ctx context.Context, args map[string]any) (*Result, error) {
	pattern, ok := args["pattern"].(string)
	if !ok || pattern == "" {
		return &Result{
			Success: false,
			Error:   "pattern is required",
		}, nil
	}

	// Build rg command
	rgArgs := []string{"-l", "-i", pattern}

	// Add include pattern if specified
	if include, ok := args["include"].(string); ok && include != "" {
		rgArgs = append(rgArgs, "--glob", include)
	}

	// Add path if specified
	if path, ok := args["path"].(string); ok && path != "" {
		rgArgs = append(rgArgs, path)
	}

	// Execute ripgrep
	cmd := exec.CommandContext(ctx, "rg", rgArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// rg returns exit code 1 when no matches found
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return &Result{
				Success: true,
				Output:  "No matches found",
				Display: formatGrepOutput(pattern, nil),
			}, nil
		}
		return &Result{
			Success: false,
			Output:  stderr.String(),
			Error:   err.Error(),
		}, nil
	}

	// Parse results
	output := stdout.String()
	matches := strings.Split(strings.TrimSpace(output), "\n")
	if len(matches) == 1 && matches[0] == "" {
		matches = nil
	}

	// Limit results
	const maxResults = 50
	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}


	return &Result{
		Success: true,
		Output:  fmt.Sprintf("Found %d file(s):\n%s", len(matches), strings.Join(matches, "\n")),
		Display: formatGrepOutput(pattern, matches),
	}, nil
}


func formatGrepOutput(pattern string, matches []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("╭─ grep: %s ", truncateString(pattern, 40)))
	sb.WriteString(strings.Repeat("─", 20))
	sb.WriteString("╮\n")

	if len(matches) == 0 {
		sb.WriteString("│ No matches found\n")
	} else {
		for _, match := range matches {
			sb.WriteString(fmt.Sprintf("│ 📄 %s\n", match))
		}
	}

	sb.WriteString(fmt.Sprintf("╰─ %d file(s) ", len(matches)))
	sb.WriteString(strings.Repeat("─", 40))
	sb.WriteString("─╯\n")

	return sb.String()
}
