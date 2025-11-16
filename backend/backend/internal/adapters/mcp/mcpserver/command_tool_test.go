package mcpserver_test

import (
	"context"
	"errors"
	"testing"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/mcp/mcpserver"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type stubExecutor struct {
	calls   [][]string
	lastDir string
	result  mcpserver.CommandResult
	err     error
}

func (s *stubExecutor) Run(_ context.Context, workingDir string, args []string) (mcpserver.CommandResult, error) {
	s.calls = append(s.calls, append([]string{}, args...))
	s.lastDir = workingDir
	return s.result, s.err
}

func TestCommandToolHandlerSuccess(t *testing.T) {
	executor := &stubExecutor{
		result: mcpserver.CommandResult{Stdout: "project created", Stderr: ""},
	}

	tool := mcpserver.NewCommandTool(
		"create",
		"Create projects",
		[]string{"create"},
		"/tmp/project",
		executor,
	)

	handler := tool.Handler.(func(context.Context, mcpserver.CommandArguments) (*mcp_golang.ToolResponse, error))

	resp, err := handler(context.Background(), mcpserver.CommandArguments{Args: []string{"demo"}})
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
		result: mcpserver.CommandResult{Stdout: "partial", Stderr: "boom"},
		err:    errors.New("command failed"),
	}

	tool := mcpserver.NewCommandTool(
		"ai",
		"AI assistant",
		[]string{"ai"},
		"/tmp/project",
		executor,
	)

	handler := tool.Handler.(func(context.Context, mcpserver.CommandArguments) (*mcp_golang.ToolResponse, error))
	resp, err := handler(context.Background(), mcpserver.CommandArguments{Args: []string{"Add"}})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "command failed")
	require.Contains(t, err.Error(), "partial")
	require.Contains(t, err.Error(), "boom")
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

	tools := mcpserver.BuildCommandTools(root, executor, "/tmp", filter)
	require.Len(t, tools, 1)
	require.Equal(t, "status", tools[0].Name)
}
