package commands

import (
	"fmt"
	"os"

	"github.com/pmaojo/kthulu-go/internal/coder"
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
		model, _ := cmd.Flags().GetString("model")
		workingDir, _ := os.Getwd()

		fmt.Println("🐙 Starting Kthulu Coder...")

// Run the coder
		messages, err := coder.Run(workingDir, model)
		if err != nil {
			return fmt.Errorf("coder error: %w", err)
		}

		// Print history
		if len(messages) > 0 {
			fmt.Println("\n📜 Chat History:")
			for _, msg := range messages {
				role := "You"
				if msg.Role != "user" {
					role = "Kthulu"
				}
				fmt.Printf("\n[%s]: %s\n", role, msg.Content)
			}
		}


		return nil
	},
}

func init() {
	coderCmd.Flags().StringP("model", "m", "gemini-2.5-flash", "AI model to use (e.g., gemini-2.5-flash, claude-3-sonnet)")
	rootCmd.AddCommand(coderCmd)
}
