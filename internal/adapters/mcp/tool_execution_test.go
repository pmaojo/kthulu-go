package mcp_test

import (
	"context"
	"testing"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/stretchr/testify/require"
)

type TestArgs struct {
	Pos1  string   `kthulu:"pos,index=0"`
	Flag1 string   `kthulu:"flag,name=foo"`
	Flag2 bool     `kthulu:"flag,name=bar"`
	Pos2  []string `kthulu:"pos,index=1,variadic"`
}

func TestNewReflectTool(t *testing.T) {
	executor := &stubExecutor{
		result: mcp.CommandResult{Stdout: "ok"},
	}

	tool := mcp.NewReflectTool[TestArgs](
		"test",
		"description",
		[]string{"base"},
		executor,
		"/tmp",
	)

	handler := tool.Handler.(func(context.Context, TestArgs) (*mcp_golang.ToolResponse, error))

	// Test Case 1: All args
	args := TestArgs{
		Pos1:  "p1",
		Flag1: "val",
		Flag2: true,
		Pos2:  []string{"v1", "v2"},
	}

	_, err := handler(context.Background(), args)
	require.NoError(t, err)

	require.Len(t, executor.calls, 1)
	call := executor.calls[0]

	// Expected: base, p1, v1, v2, --foo val, --bar
	require.Equal(t, "base", call[0])
	require.Contains(t, call, "p1")
	require.Contains(t, call, "v1")
	require.Contains(t, call, "v2")

	// Check flags
	hasFoo := false
	for i, v := range call {
		if v == "--foo" && i+1 < len(call) && call[i+1] == "val" {
			hasFoo = true
		}
	}
	require.True(t, hasFoo, "Missing --foo val")
	require.Contains(t, call, "--bar")
}
