package mcp

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// CreateMCPServerArgs are the arguments for the create_mcp_server tool.
type CreateMCPServerArgs struct {
	Name        string `json:"name" jsonschema:"required,description=Project name. A directory with this name is created under the working directory."`
	Description string `json:"description,omitempty" jsonschema:"description=Short description of what this MCP server does. Used in the generated README."`
	ModulePath  string `json:"module_path,omitempty" jsonschema:"description=Go module path for go.mod (e.g. github.com/acme/my-mcp-app). Defaults to the project name."`
}

// CreateMCPServerTool returns an MCP tool that scaffolds a new dependency-free
// MCP server project (kthulu create --template=mcp).
func CreateMCPServerTool(executor CommandExecutor, workingDir string) RegisteredTool {
	return RegisteredTool{
		Name: "create_mcp_server",
		Description: "Create a new MCP server project (dependency-free, MCP Apps / interactive UI ready). " +
			"Generates a complete Go project with a stdlib-only JSON-RPC 2.0 server, example tools, an interactive HTML dashboard view, " +
			"and passing protocol tests. Connect the binary to Claude Desktop or any MCP host. " +
			"After creation, call workdir_set with the new project directory, then run: go mod tidy && go test ./...",
		Handler: func(ctx context.Context, args CreateMCPServerArgs) (*mcp_golang.ToolResponse, error) {
			if strings.TrimSpace(args.Name) == "" {
				return nil, fmt.Errorf("name is required")
			}

			dir := resolveWorkdir(workingDir)
			cmdArgs := []string{"create", args.Name, "--template", "mcp"}
			if args.ModulePath != "" {
				cmdArgs = append(cmdArgs, "--module-path", args.ModulePath)
			}

			result, err := executor.Run(ctx, dir, cmdArgs)
			response := formatCommandResult(strings.Join(append([]string{"kthulu"}, cmdArgs...), " "), dir, result)
			if err != nil {
				return nil, fmt.Errorf("create failed: %w\n%s", err, response)
			}

			projectDir := filepath.Join(dir, args.Name)
			if _, wdErr := setSessionWorkdir(workingDir, projectDir); wdErr == nil {
				response += fmt.Sprintf("\n\n📂 Session working directory switched to %s", projectDir)
			}

			response += fmt.Sprintf(`

NEXT STEPS:
1. Finish setup:
   go mod tidy
   go test ./...
2. Build:
   go build -o %s ./cmd/%s
3. Connect to Claude Desktop (~/.config/Claude/claude_desktop_config.json):
   {"mcpServers":{"%s":{"command":"/absolute/path/to/%s"}}}
4. Extend: add tools in internal/tools/tools.go, add HTML views in internal/tools/ui/

MCP Apps (interactive UI) is wired out of the box — ask Claude to "show the status dashboard" to see it.`,
				args.Name, args.Name, args.Name, args.Name)

			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(response)), nil
		},
	}
}
