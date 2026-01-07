package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "📚 Generate API documentation",
	Long: `Generate Swagger/OpenAPI documentation for your project.
This command automatically checks for and installs the 'swag' tool if needed,
then runs 'swag init' to generate the documentation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDocCommand(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(docCmd)
}

func runDocCommand(cmd *cobra.Command, args []string) error {
	fmt.Println("📚 Generating API documentation...")

	// 1. Ensure swag is installed
	binName, err := ensureDocToolInstalled("swag", "github.com/swaggo/swag/cmd/swag@latest")
	if err != nil {
		return fmt.Errorf("failed to ensure swag is installed: %w", err)
	}

	// 2. Run swag init
	// We assume we are in the project root.
	// swag init searches for main.go in current dir by default, or we can specify it.
	// Kthulu projects usually have main in cmd/server/main.go

	mainFile := "cmd/server/main.go"
	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		// Fallback to main.go or try to find it
		if _, err := os.Stat("main.go"); err == nil {
			mainFile = "main.go"
		} else {
			// If neither exists, let swag decide or fail, but let's warn
			fmt.Println("⚠️  Could not find cmd/server/main.go or main.go. Running swag init on current directory...")
			mainFile = "."
		}
	}

	swagArgs := []string{"init"}
	if mainFile != "." {
		swagArgs = append(swagArgs, "--generalInfo", mainFile)
	}

	// Add other common flags if needed, e.g. --parseDependency
	swagArgs = append(swagArgs, "--parseDependency", "--parseInternal")

	runCmd := exec.Command(binName, swagArgs...)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	if err := runCmd.Run(); err != nil {
		return fmt.Errorf("swag init failed: %w", err)
	}

	fmt.Println("\n✅ Documentation generated successfully!")
	fmt.Println("   Open docs/swagger/index.html (if you serve it) or view docs/swagger.json")

	return nil
}

// ensureDocToolInstalled checks if a tool is installed and installs it if not.
// Duplicated/Adapted to ensure self-containment for this command.
func ensureDocToolInstalled(binName, installURL string) (string, error) {
	// Check if in PATH
	if path, err := exec.LookPath(binName); err == nil {
		return path, nil
	}

	// Determine installation path (GOBIN or GOPATH/bin)
	installPath := ""

	cmdGobin := exec.Command("go", "env", "GOBIN")
	if out, err := cmdGobin.Output(); err == nil && strings.TrimSpace(string(out)) != "" {
		installPath = strings.TrimSpace(string(out))
	} else {
		// Fallback to GOPATH/bin
		goPath := os.Getenv("GOPATH")
		if goPath == "" {
			cmd := exec.Command("go", "env", "GOPATH")
			if out, err := cmd.Output(); err == nil {
				goPath = strings.TrimSpace(string(out))
			}
		}
		if goPath == "" {
			home, _ := os.UserHomeDir()
			goPath = filepath.Join(home, "go")
		}
		installPath = filepath.Join(goPath, "bin")
	}

	candidate := filepath.Join(installPath, binName)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	// Install
	fmt.Printf("⚠️  %s not found. Installing...\n", binName)
	cmd := exec.Command("go", "install", installURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to install %s: %w", binName, err)
	}
	return candidate, nil
}
