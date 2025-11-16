package mcp_test

import (
	"context"
	"testing"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/mcp"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type noopExecutor struct{}

func (noopExecutor) Run(_ context.Context, _ string, _ []string) (mcp.CommandResult, error) {
	return mcp.CommandResult{}, nil
}

func TestServerBuilderBuildServer(t *testing.T) {
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
	require.Equal(t, 6, len(instance.Tools))
}

func TestBuildTransportHTTP(t *testing.T) {
	transport, endpoint, err := mcp.BuildTransport(mcp.TransportOptions{Kind: "http", ListenAddr: "127.0.0.1:9090", HTTPPath: "mcp"})
	require.NoError(t, err)
	require.NotNil(t, transport)
	require.Equal(t, "http://127.0.0.1:9090/mcp", endpoint)
}

func TestRegisteredToolsHaveValidSchemas(t *testing.T) {
	root := &cobra.Command{Use: "kthulu"}
	root.AddCommand(&cobra.Command{Use: "status", Run: func(cmd *cobra.Command, args []string) {}})
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
