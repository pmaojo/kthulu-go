package mcp_test

import (
	"context"
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
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

func TestBuildCommandToolsHonorsFilter(t *testing.T) {
	// Mock the registry because BuildCommandTools now uses the registry instead of walking the root command passed to it.
	original := mcp.GeneratedToolRegistry
	defer func() { mcp.GeneratedToolRegistry = original }()

	executor := &stubExecutor{}

	// Define mock tools
	mcp.GeneratedToolRegistry = map[string]func(mcp.CommandExecutor, string) mcp.RegisteredTool{
		"status": func(e mcp.CommandExecutor, wd string) mcp.RegisteredTool {
			return mcp.RegisteredTool{Name: "status"}
		},
		"deploy_apply": func(e mcp.CommandExecutor, wd string) mcp.RegisteredTool {
			return mcp.RegisteredTool{Name: "deploy_apply"}
		},
	}

	filter := func(path []string) bool {
		// path for "deploy_apply" is ["deploy", "apply"] (via split)
		return !(len(path) >= 2 && path[0] == "deploy" && path[1] == "apply")
	}

	// The root argument is now ignored by BuildCommandTools, but we pass nil or whatever.
	tools := mcp.BuildCommandTools(nil, executor, "/tmp", filter)

	require.Len(t, tools, 1)
	require.Equal(t, "status", tools[0].Name)
}
