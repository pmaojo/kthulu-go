package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/mcp"
)

var (
	mcpWorkingDir string
	mcpTransport  string
	mcpListenAddr string
	mcpHTTPPath   string
	mcpAllowList  []string
	mcpDenyList   []string
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Expose the Kthulu CLI as a Model Context Protocol server",
	Long: `Start an MCP server so AI agents can call kthulu commands like create, add, generate, ai, and more.
Use --working-dir to point the server at an existing project and --transport=http for editor or remote integrations.`,
	RunE: runMCPServer,
}

func init() {
	mcpCmd.Flags().StringVar(&mcpWorkingDir, "working-dir", "", "Working directory for executed CLI commands (default: current directory)")
	mcpCmd.Flags().StringVar(&mcpTransport, "transport", "stdio", "Transport for MCP server: stdio or http")
	mcpCmd.Flags().StringVar(&mcpListenAddr, "listen", ":8080", "Listen address when using the HTTP transport")
	mcpCmd.Flags().StringVar(&mcpHTTPPath, "http-path", "/mcp", "HTTP path for MCP requests when transport=http")
	mcpCmd.Flags().StringSliceVar(&mcpAllowList, "allow", nil, "Whitelist of CLI command paths (e.g. 'migrate up'). When set, only these commands are exposed")
	mcpCmd.Flags().StringSliceVar(&mcpDenyList, "deny", nil, "Blacklist of CLI command paths (e.g. 'deploy apply'). Denials override allows")
	rootCmd.AddCommand(mcpCmd)
}

func runMCPServer(cmd *cobra.Command, _ []string) error {
	workingDir, err := resolveWorkingDir(mcpWorkingDir)
	if err != nil {
		return err
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve kthulu binary: %w", err)
	}

	executor := mcp.NewBinaryCommandExecutor(execPath)
	tagParser := parser.NewTagParser(nil)
	builder := mcp.NewServerBuilder(mcp.ServerBuilderDependencies{
		RootCmd:   rootCmd,
		Executor:  executor,
		TagParser: tagParser,
	})
	instructions := "Expose kthulu CLI commands safely. Always respect the working directory and never run destructive shell commands outside of the provided tools."
	instance, err := builder.BuildServer(mcp.ServerOptions{
		WorkingDir: workingDir,
		AllowList:  mcpAllowList,
		DenyList:   mcpDenyList,
		Transport: mcp.TransportOptions{
			Kind:       mcpTransport,
			ListenAddr: mcpListenAddr,
			HTTPPath:   mcpHTTPPath,
		},
		Name:         "Kthulu CLI MCP",
		Version:      version,
		Instructions: instructions,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Started MCP server (%s) with %d tools in %s\n", instance.Endpoint, len(instance.Tools), workingDir)
	return instance.Server.Serve()
}

func resolveWorkingDir(flagValue string) (string, error) {
	if flagValue == "" {
		return os.Getwd()
	}

	dir, err := filepath.Abs(flagValue)
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("working directory unavailable: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}

	return dir, nil
}
