package mcpserver_test

import (
	"context"
	"errors"
	"testing"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/mcpserver"
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
