package tools

import (
	"context"
	"fmt"
	"strings"
)

// ThinkTool allows the AI to reason through complex problems
var ThinkTool = &Tool{
	Name: "think",
	Description: `Use this tool to think through complex problems.

This tool doesn't execute anything - it simply logs your thought process for transparency.

Use it when:
1. Debugging a complex issue and need to organize hypotheses
2. Planning a multi-step refactoring
3. Weighing different implementation approaches
4. Brainstorming solutions to a problem

The thought will be displayed to the user for transparency.`,
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"thought": map[string]interface{}{
				"type":        "string",
				"description": "Your thought process or reasoning",
			},
		},
		"required": []string{"thought"},
	},
	NeedsApproval: false,
	Execute:       executeThink,
}

func executeThink(ctx context.Context, args map[string]any) (*Result, error) {
	thought, ok := args["thought"].(string)
	if !ok || thought == "" {
		return &Result{
			Success: false,
			Error:   "thought is required",
		}, nil
	}

	return &Result{
		Success: true,
		Output:  "Thought logged.",
		Display: formatThinkOutput(thought),
	}, nil
}

func formatThinkOutput(thought string) string {
	var sb strings.Builder

	sb.WriteString("╭─ 💭 Thinking ")
	sb.WriteString(strings.Repeat("─", 45))
	sb.WriteString("╮\n")

	// Wrap long lines
	lines := strings.Split(thought, "\n")
	for _, line := range lines {
		sb.WriteString(fmt.Sprintf("│ %s\n", line))
	}

	sb.WriteString("╰" + strings.Repeat("─", 60) + "╯\n")

	return sb.String()
}
