package mcpserver

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/spf13/cobra"
)

// CommandArguments captures the arguments passed to a CLI-backed MCP tool.
type CommandArguments struct {
	Args []string `json:"args" jsonschema:"description=Arguments passed to the CLI command after the command name"`
}

// CommandResult contains the captured output of a CLI execution.
type CommandResult struct {
	Stdout string
	Stderr string
}

// CommandExecutor executes CLI commands.
type CommandExecutor interface {
	Run(ctx context.Context, workingDir string, args []string) (CommandResult, error)
}

// BinaryCommandExecutor executes commands using the current kthulu binary.
type BinaryCommandExecutor struct {
	binaryPath string
	extraEnv   []string
}

// NewBinaryCommandExecutor builds an executor that shells out to the kthulu binary.
func NewBinaryCommandExecutor(binaryPath string, env ...string) *BinaryCommandExecutor {
	return &BinaryCommandExecutor{binaryPath: binaryPath, extraEnv: env}
}

// Run executes the kthulu binary with the provided arguments.
func (e *BinaryCommandExecutor) Run(ctx context.Context, workingDir string, args []string) (CommandResult, error) {
	cmd := exec.CommandContext(ctx, e.binaryPath, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	if len(e.extraEnv) > 0 {
		cmd.Env = append(os.Environ(), e.extraEnv...)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return CommandResult{Stdout: stdout.String(), Stderr: stderr.String()}, err
}

// RegisteredTool represents a tool that can be registered with the MCP server.
type RegisteredTool struct {
	Name        string
	Description string
	Handler     any
}

// NewCommandTool builds a tool backed by a Cobra command path.
func NewCommandTool(name string, description string, baseArgs []string, workingDir string, executor CommandExecutor) RegisteredTool {
	handler := func(ctx context.Context, arguments CommandArguments) (*mcp_golang.ToolResponse, error) {
		resolvedArgs := append([]string{}, baseArgs...)
		resolvedArgs = append(resolvedArgs, arguments.Args...)

		result, err := executor.Run(ctx, workingDir, resolvedArgs)
		commandLabel := strings.Join(append([]string{"kthulu"}, resolvedArgs...), " ")
		response := formatCommandResult(commandLabel, workingDir, result)
		if err != nil {
			return nil, fmt.Errorf("%s failed: %w\nSTDOUT:\n%s\nSTDERR:\n%s", commandLabel, err, normalizeOutput(result.Stdout), normalizeOutput(result.Stderr))
		}

		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(response)), nil
	}

	return RegisteredTool{Name: name, Description: description, Handler: handler}
}

// BuildCommandTools converts the runnable Cobra commands into MCP tools.
func BuildCommandTools(root *cobra.Command, executor CommandExecutor, workingDir string) []RegisteredTool {
	var tools []RegisteredTool
	for _, cmd := range collectRunnableCommands(root) {
		segments := commandSegments(cmd)
		if len(segments) == 0 {
			continue
		}
		name := strings.Join(segments, "_")
		description := buildDescription(cmd)
		tools = append(tools, NewCommandTool(name, description, segments, workingDir, executor))
	}

	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})

	return tools
}

func collectRunnableCommands(root *cobra.Command) []*cobra.Command {
	var result []*cobra.Command
	for _, cmd := range root.Commands() {
		if shouldSkip(cmd) {
			continue
		}

		if cmd.Runnable() {
			result = append(result, cmd)
		}

		result = append(result, collectRunnableCommands(cmd)...)
	}
	return result
}

func shouldSkip(cmd *cobra.Command) bool {
	if cmd == nil {
		return true
	}
	if cmd.Hidden {
		return true
	}
	switch cmd.Name() {
	case "help", "completion", "mcp":
		return true
	}
	return false
}

func commandSegments(cmd *cobra.Command) []string {
	var segments []string
	current := cmd
	for current != nil {
		parent := current.Parent()
		if parent == nil {
			break
		}
		segments = append([]string{current.Name()}, segments...)
		current = parent
	}
	return segments
}

func buildDescription(cmd *cobra.Command) string {
	var builder strings.Builder
	builder.WriteString(strings.TrimSpace(cmd.Short))
	if use := strings.TrimSpace(cmd.UseLine()); use != "" {
		builder.WriteString("\nUsage: ")
		builder.WriteString(use)
	}
	if example := strings.TrimSpace(cmd.Example); example != "" {
		builder.WriteString("\nExamples:\n")
		builder.WriteString(example)
	}
	return builder.String()
}

func formatCommandResult(label string, workingDir string, result CommandResult) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Command: %s\nWorking directory: %s\n", label, workingDir))
	builder.WriteString("\nSTDOUT:\n")
	builder.WriteString(normalizeOutput(result.Stdout))
	if trimmed := strings.TrimSpace(result.Stderr); trimmed != "" {
		builder.WriteString("\n\nSTDERR:\n")
		builder.WriteString(trimmed)
	}
	return builder.String()
}

func normalizeOutput(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "<empty>"
	}
	return trimmed
}
