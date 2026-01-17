package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var coderCmd = &cobra.Command{
	Use:   "coder",
	Short: "🐙 AI-powered coding assistant with a beautiful TUI",
	Long: `Kthulu Coder is a native AI coding assistant built with Go and Bubble Tea.

Features:
  - Split-pane TUI with chat and context views
  - Multi-model support via LiteLLM (Gemini, Claude, GPT, Ollama)
  - Native tool execution (bash, file ops, search)
  - Kthulu CLI integration for scaffolding

Controls:
  Tab       - Cycle focus between panes
  Enter     - Send message
  Ctrl+L    - Clear chat
  Ctrl+O    - Open file picker
  ?         - Show help
  Ctrl+C    - Quit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Auto-load .env file if present
		_ = godotenv.Load()

		// 1. Check if 'crush' is installed
		crushPath, err := exec.LookPath("crush")
		if err != nil {
			fmt.Println("❌ 'crush' not found in PATH.")
			fmt.Println("\nKthulu Coder now uses Crush as its high-performance TUI engine.")
			fmt.Println("Please install it with:")
			fmt.Println("\n    brew install charmbracelet/tap/crush")
			fmt.Println("\nOr visit https://github.com/charmbracelet/crush for more options.")
			return nil
		}

		// 2. Ensure .kthulu directory exists
		workingDir, _ := os.Getwd()
		kthuluDir := filepath.Join(workingDir, ".kthulu")
		if err := os.MkdirAll(kthuluDir, 0755); err != nil {
			return fmt.Errorf("failed to create .kthulu directory: %w", err)
		}

		// 3. Generate/Update crush.json with Kthulu MCP
		configPath := filepath.Join(kthuluDir, "crush.json")
		if err := ensureCrushConfig(configPath); err != nil {
			return fmt.Errorf("failed to configure crush: %w", err)
		}

		fmt.Printf("🐙 Starting Kthulu Coder (via Crush) in %s\n", workingDir)

		// 4. Launch crush
		crushCmd := exec.Command(crushPath, "-c", configPath)
		crushCmd.Stdin = os.Stdin
		crushCmd.Stdout = os.Stdout
		crushCmd.Stderr = os.Stderr

		if err := crushCmd.Run(); err != nil {
			// Crush usually exits with 0 on quit, but handle errors
			return fmt.Errorf("crush exited with error: %w", err)
		}

		return nil
	},
}

func ensureCrushConfig(path string) error {
	// Default Kthulu config for Crush
	config := map[string]any{
		"$schema": "https://charm.land/crush.json",
		"mcp": map[string]any{
			"kthulu": map[string]any{
				"type":    "stdio",
				"command": "kthulu",
				"args":    []string{"mcp"},
			},
		},
		"options": map[string]any{
			"instructions": "You are the Kthulu Coder agent. Use the 'kthulu' MCP tools to manage this project's architecture. Prefer using kthulu commands for scaffolding and project analysis.",
		},
	}

	// If file exists, we could merge, but for now we'll just ensure Kthulu is there
	// unless the user has customized it heavily.
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func init() {
	coderCmd.Flags().StringP("model", "m", "gemini-2.5-flash", "AI model to use (e.g., gemini-2.5-flash, claude-3-sonnet)")
	rootCmd.AddCommand(coderCmd)
}
