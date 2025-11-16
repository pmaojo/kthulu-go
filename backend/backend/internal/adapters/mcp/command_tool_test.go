package mcp_test

import (
	"context"
	"errors"
	"testing"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/mcp"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type stubExecutor struct {
	calls   [][]string
	lastDir string
	result  mcp.CommandResult
	err     error
}

func (s *stubExecutor) Run(_ context.Context, workingDir string, args []string) (mcp.CommandResult, error) {
	s.calls = append(s.calls, append([]string{}, args...))
	s.lastDir = workingDir
	return s.result, s.err
}

func TestCommandToolHandlerSuccess(t *testing.T) {
	executor := &stubExecutor{
		result: mcp.CommandResult{Stdout: "project created", Stderr: ""},
	}

	tool := mcp.NewCommandTool(
		"create",
		"Create projects",
		[]string{"create"},
		"/tmp/project",
		executor,
	)

	handler := tool.Handler.(func(context.Context, mcp.CommandArguments) (*mcp_golang.ToolResponse, error))

	resp, err := handler(context.Background(), mcp.CommandArguments{Arg1: "demo"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Content, 1)
	require.NotNil(t, resp.Content[0].TextContent)
	require.Contains(t, resp.Content[0].TextContent.Text, "project created")
	require.Equal(t, [][]string{{"create", "demo"}}, executor.calls)
	require.Equal(t, "/tmp/project", executor.lastDir)
}

func TestCommandToolHandlerErrorIncludesOutput(t *testing.T) {
	executor := &stubExecutor{
		result: mcp.CommandResult{Stdout: "partial", Stderr: "boom"},
		err:    errors.New("command failed"),
	}

	tool := mcp.NewCommandTool(
		"ai",
		"AI assistant",
		[]string{"ai"},
		"/tmp/project",
		executor,
	)

	handler := tool.Handler.(func(context.Context, mcp.CommandArguments) (*mcp_golang.ToolResponse, error))
	resp, err := handler(context.Background(), mcp.CommandArguments{Arg1: "Add"})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "command failed")
	require.Contains(t, err.Error(), "partial")
	require.Contains(t, err.Error(), "boom")
}

func TestCommandArgumentsResolveArgs(t *testing.T) {
	arguments := mcp.CommandArguments{
		Arg1:     "one",
		Arg3:     "two",
		ArgsText: "three four",
		ArgsJSON: `["five","six"]`,
	}
	resolved, err := arguments.ResolveArgs()
	require.NoError(t, err)
	require.Equal(t, []string{"one", "two", "three", "four", "five", "six"}, resolved)
}

func TestCommandArgumentsResolveArgsJSONError(t *testing.T) {
	arguments := mcp.CommandArguments{ArgsJSON: "not-json"}
	resolved, err := arguments.ResolveArgs()
	require.Error(t, err)
	require.Nil(t, resolved)
}

func TestBuildCommandToolsHonorsFilter(t *testing.T) {
	executor := &stubExecutor{}
	root := &cobra.Command{Use: "kthulu"}

	statusCmd := &cobra.Command{Use: "status", Run: func(cmd *cobra.Command, args []string) {}}
	root.AddCommand(statusCmd)

	destructive := &cobra.Command{Use: "deploy"}
	apply := &cobra.Command{Use: "apply", Run: func(cmd *cobra.Command, args []string) {}}
	destructive.AddCommand(apply)
	root.AddCommand(destructive)

	filter := func(path []string) bool {
		return !(len(path) >= 2 && path[0] == "deploy" && path[1] == "apply")
	}

	tools := mcp.BuildCommandTools(root, executor, "/tmp", filter)
	require.Len(t, tools, 1)
	require.Equal(t, "status", tools[0].Name)
}
