package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pmaojo/kthulu-go/internal/adapters/cli/debugui"
)

var (
	debugCmdStr   string
	debugPersist  bool
	debugTestWatch bool
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "🐞 Interactive debug daemon and runtime monitor",
	Long: `Starts the Kthulu Debug Daemon TUI.

This command runs your application as a subprocess and provides a real-time
interactive dashboard (like Laravel Telescope) in your terminal.

It captures logs to display:
  - 🌐 HTTP Requests (Latency, Status, Method)
  - 💾 Database Queries (SQL, Duration)
  - 📝 Raw Application Logs
  - 📈 System Resources

Examples:
  kthulu debug
  kthulu debug --cmd="go run cmd/server/main.go"
  kthulu debug --persist
  kthulu debug --test-watch
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine command to run
		runCmd := debugCmdStr
		if runCmd == "" {
			// Auto-detect standard Kthulu entrypoint
			if _, err := os.Stat("cmd/server/main.go"); err == nil {
				runCmd = "go run cmd/server/main.go"
			} else if _, err := os.Stat("main.go"); err == nil {
				runCmd = "go run main.go"
			} else {
				// If not found, and we are not in test watch mode, error out.
				// For test watch mode, running the app is optional/secondary.
				if !debugTestWatch {
					return fmt.Errorf("could not auto-detect entrypoint. Please use --cmd")
				}
				runCmd = "echo 'No app running'"
			}
		}

		// Handle persistence override via Env
		if os.Getenv("KTHULU_DEBUG_PERSIST") == "true" {
			debugPersist = true
		}

		// Split command string into name and args
		parts := strings.Fields(runCmd)
		if len(parts) == 0 {
			return fmt.Errorf("invalid command")
		}

		// Create the subprocess command
		subProc := exec.Command(parts[0], parts[1:]...)
		// We do NOT attach Stdout/Stderr here because the TUI will capture them.
		// However, we set the Env to force unbuffered output if possible
		subProc.Env = append(os.Environ(), "PYTHONUNBUFFERED=1", "GOLANG_LOG=text")

		// Initialize the Bubble Tea model
		model := debugui.NewModel(subProc, debugPersist, debugTestWatch)

		// Run the TUI
		p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running debug daemon: %w", err)
		}

		return nil
	},
}

func init() {
	debugCmd.Flags().StringVar(&debugCmdStr, "cmd", "", "Command to run the application (default: auto-detect 'go run ...')")
	debugCmd.Flags().BoolVar(&debugPersist, "persist", false, "Persist debug events to disk")
	debugCmd.Flags().BoolVar(&debugTestWatch, "test-watch", false, "Watch tests and display status")
	rootCmd.AddCommand(debugCmd)
}
