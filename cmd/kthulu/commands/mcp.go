package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
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

	executor := mcp.NewBinaryCommandExecutor(execPath, "KTHULU_MCP_MODE=1")
	tagParser := parser.NewTagParser(nil)
	builder := mcp.NewServerBuilder(mcp.ServerBuilderDependencies{
		RootCmd:   rootCmd,
		Executor:  executor,
		TagParser: tagParser,
	})
	instructions := `Kthulu generates production-ready Go applications. Follow this workflow:

CREATING AN APPLICATION (the golden path):
1. Model the domain first: list every entity with its REAL fields, validation
   rules and relations (name:type[:rules]; belongs_to for associations).
   Optionally check the model with review_domain_model.
2. Call scaffold_project with that model. NEVER create an app whose entities
   only have a name field - that means the domain was not modeled.
3. scaffold_project switches the session working directory to the new project
   automatically. Finish setup via shell_execute:
   go run github.com/a-h/templ/cmd/templ@v0.3.977 generate ./... && go mod tidy
4. Verify with go_build, then go_test.

EVOLVING AN APPLICATION:
- add_module <name> <field:type:rules...> for new entities (always pass fields)
- migrate diff after changing entity structs to generate the SQL migration
- workdir_get / workdir_set to orient yourself or switch projects

Also available: file editing (fs_*), code search (code_search, file_glob),
Go AST analysis (go_outline, go_find_symbol, go_symbol_source), database
introspection (db_schema, db_query), the Go toolchain (go_test, go_build,
go_vet) and file watching (watch_*). Always respect the working directory and
never run destructive shell commands outside of the provided tools.`
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

	fmt.Fprintf(cmd.ErrOrStderr(), "Started MCP server (%s) with %d tools in %s\n", instance.Endpoint, len(instance.Tools), workingDir)
	if err := instance.Server.Serve(); err != nil {
		return err
	}

	if instance.Done != nil {
		<-instance.Done
	}

	return nil
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
