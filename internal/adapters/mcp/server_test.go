package mcp_test

import (
	"context"
	"testing"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type noopExecutor struct{}

func (noopExecutor) Run(_ context.Context, _ string, _ []string) (mcp.CommandResult, error) {
	return mcp.CommandResult{}, nil
}

// Named struct for testing to avoid jsonschema issues with anonymous structs
type StatusArgs struct {
	Force bool `json:"force"`
}

func TestServerBuilderBuildServer(t *testing.T) {
	// Mock registry
	original := mcp.GeneratedToolRegistry
	defer func() { mcp.GeneratedToolRegistry = original }()
	mcp.GeneratedToolRegistry = map[string]func(mcp.CommandExecutor, string) mcp.RegisteredTool{
		"status": func(e mcp.CommandExecutor, wd string) mcp.RegisteredTool {
			return mcp.RegisteredTool{
				Name: "status",
				Handler: func(ctx context.Context, args StatusArgs) (*mcp_golang.ToolResponse, error) {
					return nil, nil
				},
			}
		},
	}

	root := &cobra.Command{Use: "kthulu"}
	root.AddCommand(&cobra.Command{Use: "status", Short: "Check status", Run: func(cmd *cobra.Command, args []string) {}})

	builder := mcp.NewServerBuilder(mcp.ServerBuilderDependencies{
		RootCmd:   root,
		Executor:  noopExecutor{},
		TagParser: parser.NewTagParser(nil),
	})

	instance, err := builder.BuildServer(mcp.ServerOptions{
		WorkingDir: "/tmp", // tests do not touch disk
		Transport:  mcp.TransportOptions{Kind: "stdio"},
		Name:       "Test MCP",
		Version:    "dev",
	})
	require.NoError(t, err)
	require.NotNil(t, instance.Server)
	require.Equal(t, "stdio", instance.Endpoint)
	require.Equal(t, 13, len(instance.Tools))
}

func TestBuildTransportHTTP(t *testing.T) {
	transport, endpoint, done, err := mcp.BuildTransport(mcp.TransportOptions{Kind: "http", ListenAddr: "127.0.0.1:9090", HTTPPath: "mcp"})
	require.NoError(t, err)
	require.NotNil(t, transport)
	require.Nil(t, done)
	require.Equal(t, "http://127.0.0.1:9090/mcp", endpoint)
}

func TestRegisteredToolsHaveValidSchemas(t *testing.T) {
	original := mcp.GeneratedToolRegistry
	defer func() { mcp.GeneratedToolRegistry = original }()
	mcp.GeneratedToolRegistry = map[string]func(mcp.CommandExecutor, string) mcp.RegisteredTool{
		"status": func(e mcp.CommandExecutor, wd string) mcp.RegisteredTool {
			return mcp.RegisteredTool{
				Name: "status",
				Handler: func(ctx context.Context, args StatusArgs) (*mcp_golang.ToolResponse, error) {
					return nil, nil
				},
			}
		},
	}

	root := &cobra.Command{Use: "kthulu"}
	factory := mcp.NewToolFactory(root, noopExecutor{}, parser.NewTagParser(nil))
	tools := factory.BuildTools("/tmp", mcp.NewAllowDenyFilter(nil, nil))
	for _, tool := range tools {
		tool := tool
		t.Run(tool.Name, func(t *testing.T) {
			server := mcp_golang.NewServer(stdio.NewStdioServerTransport())
			require.NotPanics(t, func() {
				require.NoError(t, server.RegisterTool(tool.Name, tool.Description, tool.Handler))
			})
		})
	}
}
