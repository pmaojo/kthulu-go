package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// KthuluTool executes kthulu CLI commands
var KthuluTool = &Tool{
	Name: "kthulu",
	Description: `Execute Kthulu CLI commands for project scaffolding and analysis.

Available commands:
- kthulu add module <name> [fields...] - Create a new module
- kthulu add component <type> <name> - Add a component (handler, service, repository)
- kthulu analyze - Analyze project structure
- kthulu status - Show project health

Use this tool to leverage Kthulu's scaffolding and code generation capabilities.`,
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The kthulu subcommand to execute (e.g., 'add module user name:string')",
			},
		},
		"required": []string{"command"},
	},
	NeedsApproval: true,
	Execute:       executeKthulu,
}

func executeKthulu(ctx context.Context, args map[string]any) (*Result, error) {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return &Result{
			Success: false,
			Error:   "command is required",
		}, nil
	}

	// Split command into args
	cmdArgs := strings.Fields(command)

	// Execute kthulu command
	cmd := exec.CommandContext(ctx, "kthulu", cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	if err != nil {
		return &Result{
			Success: false,
			Output:  output,
			Display: formatKthuluOutput(command, output, false),
			Error:   err.Error(),
		}, nil
	}

	return &Result{
		Success: true,
		Output:  output,
		Display: formatKthuluOutput(command, output, true),
	}, nil
}

func formatKthuluOutput(command, output string, success bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("╭─ 🐙 kthulu %s ", truncateString(command, 40)))
	sb.WriteString(strings.Repeat("─", 15))
	sb.WriteString("╮\n")

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line != "" {
			sb.WriteString(fmt.Sprintf("│ %s\n", line))
		}
	}

	if success {
		sb.WriteString("╰" + strings.Repeat("─", 50) + " ✓ Done ─╯\n")
	} else {
		sb.WriteString("╰" + strings.Repeat("─", 50) + " ✗ Error ─╯\n")
	}

	return sb.String()
}
