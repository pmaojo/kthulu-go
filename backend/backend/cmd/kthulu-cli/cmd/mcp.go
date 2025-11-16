package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
	mcptransport "github.com/metoro-io/mcp-golang/transport"
	mcphttp "github.com/metoro-io/mcp-golang/transport/http"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/mcpserver"
	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/parser"
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

	executor := mcpserver.NewBinaryCommandExecutor(execPath)
	tagParser := parser.NewTagParser(nil)
	factory := mcpserver.NewToolFactory(rootCmd, executor, tagParser)
	filter := mcpserver.NewAllowDenyFilter(mcpAllowList, mcpDenyList)
	tools := factory.BuildTools(workingDir, filter)

	transport, endpoint, err := buildMCPTransport(mcpTransport, mcpListenAddr, mcpHTTPPath)
	if err != nil {
		return err
	}
	instructions := "Expose kthulu CLI commands safely. Always respect the working directory and never run destructive shell commands outside of the provided tools."

	server := mcp_golang.NewServer(
		transport,
		mcp_golang.WithName("Kthulu CLI MCP"),
		mcp_golang.WithVersion(version),
		mcp_golang.WithInstructions(instructions),
	)

	for _, tool := range tools {
		if err := server.RegisterTool(tool.Name, tool.Description, tool.Handler); err != nil {
			return fmt.Errorf("register tool %s: %w", tool.Name, err)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Started MCP server (%s) with %d tools in %s\n", endpoint, len(tools), workingDir)
	return server.Serve()
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

func buildMCPTransport(kind string, listenAddr string, httpPath string) (mcptransport.Transport, string, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "", "stdio":
		return stdio.NewStdioServerTransport(), "stdio", nil
	case "http":
		path := normalizeHTTPPath(httpPath)
		transport := mcphttp.NewHTTPTransport(path)
		addr := strings.TrimSpace(listenAddr)
		if addr != "" {
			transport = transport.WithAddr(addr)
		}
		displayAddr := renderHTTPAddress(addr)
		return transport, fmt.Sprintf("%s%s", displayAddr, path), nil
	default:
		return nil, "", fmt.Errorf("unsupported MCP transport %s", kind)
	}
}

func normalizeHTTPPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "/mcp"
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}

func renderHTTPAddress(addr string) string {
	display := strings.TrimSpace(addr)
	if display == "" {
		display = ":8080"
	}
	if strings.HasPrefix(display, "http://") || strings.HasPrefix(display, "https://") {
		return display
	}
	if strings.HasPrefix(display, ":") {
		display = "localhost" + display
	}
	return "http://" + display
}
