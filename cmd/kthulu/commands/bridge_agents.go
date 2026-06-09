package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Common MCP config structure used by Claude and others
type mcpConfig struct {
	MCPServers map[string]mcpServerConfig `json:"mcpServers"`
}

type mcpServerConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

var claudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "🤖 Configure and launch Claude CLI with Kthulu superpowers",
	Long: `Auto-configures the current project for Claude CLI by generating an .mcp.json file 
and launching the agent. Requires 'claude' to be installed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Check for binary
		claudePath, err := exec.LookPath("claude")
		if err != nil {
			fmt.Println("❌ 'claude' binary not found in PATH.")
			fmt.Println("Please install it: npm install -g @anthropic-ai/claude-code")
			return nil
		}

		// 2. Generate config
		workingDir, _ := os.Getwd()
		configPath := filepath.Join(workingDir, ".mcp.json") // Claude CLI looks for this
		if err := ensureMCPConfig(configPath); err != nil {
			return fmt.Errorf("failed to generate .mcp.json: %w", err)
		}

		fmt.Printf("🤖 Launching Claude CLI in %s...\n", workingDir)

		// 3. Launch
		c := exec.Command(claudePath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

var geminiCmd = &cobra.Command{
	Use:   "gemini",
	Short: "✨ Launch Gemini CLI with Kthulu Superpowers (Context + MCP)",
	Long: `Auto-configures and launches the Gemini CLI.
1. Generates '.gemini/settings.json' to connect to Kthulu MCP.
2. Generates 'GEMINI.md' with project architecture context (Modules, Layers).
3. Launches 'gemini' interactive session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Check for binary
		geminiPath, err := exec.LookPath("gemini")
		if err != nil {
			fmt.Println("❌ 'gemini' binary not found in PATH.")
			fmt.Println("Please install it via npm or your package manager.")
			return nil
		}

		absPath, err := os.Executable()
		if err != nil {
			return err
		}
		wd, _ := os.Getwd()

		// 2. Configure MCP (.gemini/settings.json)
		geminiDir := filepath.Join(wd, ".gemini")
		if err := os.MkdirAll(geminiDir, 0755); err == nil {
			configPath := filepath.Join(geminiDir, "settings.json")
			configureGeminiMCP(configPath, absPath)
		}

		// 3. Generate Context (GEMINI.md)
		// We could use the parser service here, but for MVP we'll inject basic Kthulu architecture info
		// In a real implementation, we would inject the output of 'kthulu analyze'
		contextPath := filepath.Join(wd, "GEMINI.md")

		contextContent := `# Kthulu Project Context

This is a Kthulu Framework project. It follows a Modular Monolith architecture.

## Architectural Rules
- **Vertical Slices**: Code is organized by business domain (Modules).
- **Layers**: 
  - ` + "`API`" + `: Transport layer (HTTP, gRPC).
  - ` + "`Core`" + `: Business logic and use cases.
  - ` + "`Store`" + `: Data access and persistence.
- **Dependencies**: Modules should be loosely coupled.

## Available Tools (Kthulu MCP)
You have access to a powerful set of tools via the Kthulu MCP server:
- **Shell**: ` + "`host_shell_execute`" + `
- **Git**: ` + "`git_status`" + `, ` + "`git_diff`" + `, ` + "`git_log`" + `
- **Files**: ` + "`fs_read`" + `, ` + "`fs_write`" + `, ` + "`fs_edit`" + ` (exact string replacement), ` + "`fs_list`" + `, ` + "`fs_move`" + `, ` + "`fs_delete`" + `
- **Search**: ` + "`code_search`" + ` (regex with context lines), ` + "`file_glob`" + ` (supports ` + "`**`" + `)
- **Go AST**: ` + "`go_outline`" + `, ` + "`go_find_symbol`" + `, ` + "`go_symbol_source`" + `
- **Database**: ` + "`db_schema`" + `, ` + "`db_query`" + ` (SQLite and PostgreSQL)
- **Go Toolchain**: ` + "`go_test`" + `, ` + "`go_build`" + `, ` + "`go_vet`" + `
- **File Watching**: ` + "`watch_start`" + `, ` + "`watch_events`" + `, ` + "`watch_stop`" + `, ` + "`watch_list`" + `
- **Architecture**: Create modules, components, and manage the graph.

Use these tools to verify the state of the project and perform tasks.

## Workflows & Troubleshooting

### Dependency Resolution in Workspaces
If you are working in a monorepo or a folder with a parent ` + "`go.work`" + ` file (e.g., ` + "`kthulu-go`" + `):
- **ALWAYS** export ` + "`GOWORK=off`" + ` when running ` + "`go run`" + ` or ` + "`go test`" + ` inside a generated project.
- **Preferred**: Use ` + "`kthulu dev`" + ` effectively, as it handles this automatically.
- If dependencies are missing ("package not in std"): run ` + "`go mod tidy`" + ` then try again with ` + "`GOWORK=off`" + `.

### Starting the Server
- **Check Ports**: Before starting the server, check if port 8080 is free: ` + "`lsof -i :8080`" + `.
- **Use the CLI**: Prefer running ` + "`kthulu dev`" + ` which starts both backend and frontend with AI self-healing.
- **Manual Start**: ` + "`cd <project> && GOWORK=off go run cmd/server/main.go`" + `

### Common Issues
- **"bind: address already in use"**: Kill the process on port 8080 or use ` + "`kthulu dev`" + `.
- **"package ... is not in std"**: You are likely mistakenly using the parent ` + "`go.work`" + `. Set ` + "`GOWORK=off`" + `.
`
		// Only write if not exists to avoid overwriting user custom notes?
		// Or maybe overwrite/prepend? Let's write for now as it's a "bridge" command.
		os.WriteFile(contextPath, []byte(contextContent), 0644)

		fmt.Printf("✨ Launching Gemini CLI in %s...\n", wd)
		fmt.Println("   - MCP Configured ✅")
		fmt.Println("   - Context Injected (GEMINI.md) ✅")

		// 4. Launch
		c := exec.Command(geminiPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func configureGeminiMCP(configPath, kthuluPath string) {
	// Load existing or create new
	var config map[string]any
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &config)
	}
	if config == nil {
		config = make(map[string]any)
	}

	// Add mcpServers/kthulu
	mcpServers, _ := config["mcpServers"].(map[string]any)
	if mcpServers == nil {
		mcpServers = make(map[string]any)
	}
	mcpServers["kthulu"] = map[string]any{
		"command": kthuluPath,
		"args":    []string{"mcp"},
	}
	config["mcpServers"] = mcpServers

	// Write back
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configPath, data, 0644)
}

func init() {
	rootCmd.AddCommand(claudeCmd)
	rootCmd.AddCommand(geminiCmd)
}

func ensureMCPConfig(path string) error {
	absPath, err := os.Executable()
	if err != nil {
		return err
	}

	config := mcpConfig{
		MCPServers: map[string]mcpServerConfig{
			"kthulu": {
				Command: absPath,
				Args:    []string{"mcp"},
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Always overwrite to ensure latest path/args
	return os.WriteFile(path, data, 0644)
}
