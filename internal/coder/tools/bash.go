package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// BashTool executes shell commands
var BashTool = &Tool{
	Name: "bash",
	Description: `Execute a bash command in the shell.

Use this tool to:
- Run build commands (go build, npm run, etc.)
- Execute tests (go test, pytest, etc.)
- Git operations (git status, git diff, git commit)
- File operations (ls, cat, find, etc.)

Important:
- Commands run in the current working directory
- Long-running commands will timeout after 30 seconds
- Avoid interactive commands that require user input
- Be careful with destructive commands (rm, etc.)`,
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The bash command to execute",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default: 30, max: 300)",
			},
		},
		"required": []string{"command"},
	},
	NeedsApproval: true,
	Execute:       executeBash,
}

func executeBash(ctx context.Context, args map[string]any) (*Result, error) {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return &Result{
			Success: false,
			Error:   "command is required",
		}, nil
	}

	// Parse timeout
	timeout := 30 * time.Second
	if t, ok := args["timeout"].(float64); ok && t > 0 {
		if t > 300 {
			t = 300
		}
		timeout = time.Duration(t) * time.Second
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute command
	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Build output
	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	// Truncate if too long
	const maxOutput = 10000
	if len(output) > maxOutput {
		output = output[:maxOutput] + "\n... (output truncated)"
	}

	// Check for errors
	if ctx.Err() == context.DeadlineExceeded {
		return &Result{
			Success: false,
			Output:  output,
			Display: formatBashOutput(command, output, -1, true),
			Error:   "command timed out",
		}, nil
	}

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return &Result{
				Success: false,
				Output:  output,
				Display: formatBashOutput(command, output, -1, false),
				Error:   err.Error(),
			}, nil
		}
	}

	return &Result{
		Success: exitCode == 0,
		Output:  output,
		Display: formatBashOutput(command, output, exitCode, false),
	}, nil
}

func formatBashOutput(command, output string, exitCode int, timedOut bool) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("╭─ bash: %s ", truncateString(command, 50)))
	sb.WriteString(strings.Repeat("─", 30))
	sb.WriteString("╮\n")

	// Output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line != "" {
			sb.WriteString(fmt.Sprintf("│ %s\n", line))
		}
	}

	// Footer
	if timedOut {
		sb.WriteString("╰" + strings.Repeat("─", 50) + " ⏱ Timeout ─╯\n")
	} else if exitCode == 0 {
		sb.WriteString("╰" + strings.Repeat("─", 50) + " ✓ Exit 0 ─╯\n")
	} else {
		sb.WriteString(fmt.Sprintf("╰"+strings.Repeat("─", 48)+" ✗ Exit %d ─╯\n", exitCode))
	}

	return sb.String()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
