package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/spf13/cobra"
)

// CommandArguments captures the arguments passed to a CLI-backed MCP tool.
//
// The schema follows the MCP server restrictions that only string fields are allowed.
// Callers can either populate the explicit argN fields, provide a space-separated
// argsText value, or include argsJSON containing a JSON array of strings.
type CommandArguments struct {
	Arg1     string `json:"arg1,omitempty" jsonschema:"description=First positional argument"`
	Arg2     string `json:"arg2,omitempty" jsonschema:"description=Second positional argument"`
	Arg3     string `json:"arg3,omitempty" jsonschema:"description=Third positional argument"`
	Arg4     string `json:"arg4,omitempty" jsonschema:"description=Fourth positional argument"`
	Arg5     string `json:"arg5,omitempty" jsonschema:"description=Fifth positional argument"`
	Arg6     string `json:"arg6,omitempty" jsonschema:"description=Sixth positional argument"`
	Arg7     string `json:"arg7,omitempty" jsonschema:"description=Seventh positional argument"`
	Arg8     string `json:"arg8,omitempty" jsonschema:"description=Eighth positional argument"`
	Arg9     string `json:"arg9,omitempty" jsonschema:"description=Ninth positional argument"`
	Arg10    string `json:"arg10,omitempty" jsonschema:"description=Tenth positional argument"`
	ArgsText string `json:"args_text,omitempty" jsonschema:"description=Optional whitespace-delimited argument list when more than ten values are required"`
	ArgsJSON string `json:"args_json,omitempty" jsonschema:"description=Optional JSON array of arguments for full fidelity"`
}

// ResolveArgs converts the provided command arguments into the final slice.
func (a CommandArguments) ResolveArgs() ([]string, error) {
	values := []string{a.Arg1, a.Arg2, a.Arg3, a.Arg4, a.Arg5, a.Arg6, a.Arg7, a.Arg8, a.Arg9, a.Arg10}
	var args []string
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			args = append(args, trimmed)
		}
	}
	if extra := strings.TrimSpace(a.ArgsText); extra != "" {
		args = append(args, strings.Fields(extra)...)
	}
	if rawJSON := strings.TrimSpace(a.ArgsJSON); rawJSON != "" {
		var decoded []string
		if err := json.Unmarshal([]byte(rawJSON), &decoded); err != nil {
			return nil, fmt.Errorf("parse args_json: %w", err)
		}
		args = append(args, decoded...)
	}
	return args, nil
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
		dynamicArgs, err := arguments.ResolveArgs()
		if err != nil {
			return nil, err
		}
		resolvedArgs = append(resolvedArgs, dynamicArgs...)

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
func BuildCommandTools(root *cobra.Command, executor CommandExecutor, workingDir string, filter CommandFilter) []RegisteredTool {
	var tools []RegisteredTool
	for _, cmd := range collectRunnableCommands(root) {
		segments := commandSegments(cmd)
		if len(segments) == 0 {
			continue
		}
		if filter != nil && !filter(segments) {
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
