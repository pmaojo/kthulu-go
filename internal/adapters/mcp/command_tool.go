package mcp

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

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
	if dir := resolveWorkdir(workingDir); dir != "" {
		cmd.Dir = dir
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

// BuildCommandTools converts the runnable Cobra commands into MCP tools.
func BuildCommandTools(root *cobra.Command, executor CommandExecutor, workingDir string, filter CommandFilter) []RegisteredTool {
	var tools []RegisteredTool

	// Use the generated registry if available
	if GeneratedToolRegistry != nil {
		for name, factory := range GeneratedToolRegistry {
			segments := strings.Split(name, "_")
			if filter != nil && !filter(segments) {
				continue
			}
			tools = append(tools, factory(executor, workingDir))
		}
	} else {
		// Fallback for when tools are not generated (should not happen if mcp-gen is run)
		// But in unit tests, registry might be empty if we don't mock it.
		// Since we deleted dynamic tool logic, we can't fallback easily without code duplication.
		// We'll assume registry is populated or we accept empty tools.
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

// runCreateCLI runs a kthulu create command and switches the session working
// directory to the newly created project. It returns the formatted output.
// errorVerb is used in the error message (e.g. "scaffold", "create").
func runCreateCLI(ctx context.Context, executor CommandExecutor, dir, workingDir, projectName, errorVerb string, cmdArgs []string) (string, error) {
	result, err := executor.Run(ctx, dir, cmdArgs)
	response := formatCommandResult(strings.Join(append([]string{"kthulu"}, cmdArgs...), " "), dir, result)
	if err != nil {
		return "", fmt.Errorf("%s failed: %w\n%s", errorVerb, err, response)
	}
	projectDir := filepath.Join(dir, projectName)
	if _, wdErr := setSessionWorkdir(workingDir, projectDir); wdErr == nil {
		response += fmt.Sprintf("\n\n📂 Session working directory switched to %s", projectDir)
	}
	return response, nil
}

func normalizeOutput(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "<empty>"
	}
	return trimmed
}
