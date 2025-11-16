package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/mcpserver"
	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/parser"
)

var (
	mcpWorkingDir string
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Expose the Kthulu CLI as a Model Context Protocol server",
	Long: `Start an MCP stdio server so AI agents can call kthulu commands like create, add, generate, ai, and more.
Use --working-dir to point the server at an existing project.`,
	RunE: runMCPServer,
}

func init() {
	mcpCmd.Flags().StringVar(&mcpWorkingDir, "working-dir", "", "Working directory for executed CLI commands (default: current directory)")
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
	tools := mcpserver.BuildCommandTools(rootCmd, executor, workingDir)

	tagParser := parser.NewTagParser(nil)
	guideTool := mcpserver.NewGuideTaggingService(tagParser).Tool(workingDir)
	tools = append(tools, guideTool)

	insights := mcpserver.NewProjectInsightsService(tagParser)
	tools = append(tools,
		insights.OverviewTool(workingDir),
		insights.ModulesTool(workingDir),
		insights.TagsTool(workingDir),
		insights.DependenciesTool(workingDir),
	)

	transport := stdio.NewStdioServerTransport()
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

	fmt.Fprintf(cmd.OutOrStdout(), "Started MCP server with %d tools in %s\n", len(tools), workingDir)
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
