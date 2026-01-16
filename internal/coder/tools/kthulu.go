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
	Description: `Execute Kthulu CLI commands for project scaffolding, AI assistance, and development.

## Scaffolding & Code Generation
- kthulu create <name> - Create a new Kthulu project
- kthulu add module <name> [fields...] - Create a new module with optional fields
- kthulu add component <type> <name> - Add component (handler, service, repository)
- kthulu add auth - Add JWT authentication module
- kthulu generate <type> <name> - Generate code artifacts

## AI Commands
- kthulu ai <prompt> - AI-powered code assistance
- kthulu ai review - AI code review and fixes
- kthulu ai optimize - AI performance optimization
- kthulu ai scaffold <desc> - AI project scaffolding
- kthulu ai gen-feature <desc> - Generate BDD feature file
- kthulu ai gen-steps <file> - Generate step definitions

## BDD Testing
- kthulu bdd features - List all feature files
- kthulu bdd run - Run BDD scenarios

## Database
- kthulu migrate up - Apply migrations
- kthulu migrate down - Revert migration
- kthulu migrate create <name> - Create migration
- kthulu migrate status - Show DB version

## Project Analysis
- kthulu analyze - Analyze project structure
- kthulu status - Show project health
- kthulu doctor - Diagnose environment
- kthulu audit - Security/compliance audit

## Development
- kthulu dev - Start dev server with self-healing
- kthulu deploy - Zero-config deployment
- kthulu doc - Generate API documentation

Use this tool to leverage Kthulu's full capabilities.`,
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The kthulu subcommand to execute (e.g., 'add module user name:string', 'ai review', 'status')",
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
