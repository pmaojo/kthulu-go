package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version information
var (
	version = "dev"
	build   = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "kthulu",
	Short: "🚀 Kthulu Framework CLI - The Ultimate Go Development Experience",
	Long: `
🦑 Kthulu Framework CLI v` + version + `

The most powerful Go framework CLI with AI-powered code generation,
enterprise security, and zero-config deployment.

🚀 Features:
  • AI-guided project creation
  • Smart dependency resolution  
  • Enterprise security built-in
  • Multi-cloud deployment
  • Real-time collaboration

💡 Get started:
  kthulu create my-app --ai-guided
  kthulu ai "Add Stripe payments to my API"
  kthulu deploy --cloud=aws --scale=auto
`,
	Version: version,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Core commands
	rootCmd.AddCommand(newCmd) // kthulu create/new
	rootCmd.AddCommand(addCmd) // kthulu add module

	// AI commands
	rootCmd.AddCommand(aiCmd) // kthulu ai "prompt"

	// Enterprise commands
	rootCmd.AddCommand(auditCmd)   // kthulu audit
	rootCmd.AddCommand(deployCmd)  // kthulu deploy
	rootCmd.AddCommand(statusCmd)  // kthulu status
	rootCmd.AddCommand(upgradeCmd) // kthulu upgrade
	rootCmd.AddCommand(secureCmd)  // kthulu secure

	// Other commands
	rootCmd.AddCommand(migrateCmd) // kthulu migrate
}
