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

var (
	docDir         string
	docGeneralInfo string
)

func init() {
	docCmd.Flags().StringVar(&docDir, "dir", "", "Directory to search for main.go (default: project root)")
	docCmd.Flags().StringVar(&docGeneralInfo, "generalInfo", "", "API general info file (default: cmd/server/main.go)")
	rootCmd.AddCommand(docCmd)
}

func runDocCommand(cmd *cobra.Command, args []string) error {
	fmt.Println("📚 Generating API documentation...")

	projectRoot := ""
	var err error

	// 0. Determine project root / working directory
	if docDir != "" {
		projectRoot = docDir
		fmt.Printf("📂 Using specified directory: %s\n", projectRoot)
	} else {
		// Find project root automatically
		projectRoot, err = findProjectRoot()
		if err != nil {
			// Fallback to current directory if not found, but warn
			fmt.Println("⚠️  Could not find go.mod, assuming current directory is project root.")
			projectRoot, _ = os.Getwd()
		} else {
			fmt.Printf("📂 Found project root: %s\n", projectRoot)
		}
	}

	// 1. Ensure swag is installed
	binName, err := ensureDocToolInstalled("swag", "github.com/swaggo/swag/cmd/swag@latest")
	if err != nil {
		return fmt.Errorf("failed to ensure swag is installed: %w", err)
	}

	// 2. Determine main file
	mainFile := "cmd/server/main.go" // Default assumption

	if docGeneralInfo != "" {
		mainFile = docGeneralInfo
		fmt.Printf("📄 Using specified general info: %s\n", mainFile)
	} else {
		// Auto-detect main file
		if _, err := os.Stat(filepath.Join(projectRoot, mainFile)); os.IsNotExist(err) {
			// Fallback to main.go or try to find it
			if _, err := os.Stat(filepath.Join(projectRoot, "main.go")); err == nil {
				mainFile = "main.go"
			} else if _, err := os.Stat(filepath.Join(projectRoot, "cmd/kthulu-cli/main.go")); err == nil {
				mainFile = "cmd/kthulu-cli/main.go"
			} else {
				// If neither exists, let swag decide or fail, but let's warn
				fmt.Println("⚠️  Could not find cmd/server/main.go, main.go or cmd/kthulu-cli/main.go. Running swag init on current directory...")
				mainFile = "."
			}
		}
	}

	swagArgs := []string{"init"}
	if mainFile != "." {
		swagArgs = append(swagArgs, "--generalInfo", mainFile)
	}

	// Use projectRoot as dir if not ".", or if explicit dir was passed
	if projectRoot != "" {
		swagArgs = append(swagArgs, "--dir", projectRoot)
	}

	// Add other common flags if needed, e.g. --parseDependency
	swagArgs = append(swagArgs, "--parseDependency", "--parseInternal")

	runCmd := exec.Command(binName, swagArgs...)
	// runCmd.Dir = projectRoot // Swag takes --dir arg, so we can run from anywhere or projectRoot. 
	// Running from projectRoot is safer for relative paths in generalInfo if strictly relative.
	if projectRoot != "" {
		runCmd.Dir = projectRoot
	} else {
		runCmd.Dir, _ = os.Getwd()
	}
	
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
	// 1. Check if in PATH
	if path, err := exec.LookPath(binName); err == nil {
		return path, nil
	}

	// 2. Check explicitly in GOBIN
	gobin := os.Getenv("GOBIN")
	if gobin == "" {
		cmd := exec.Command("go", "env", "GOBIN")
		if out, err := cmd.Output(); err == nil && strings.TrimSpace(string(out)) != "" {
			gobin = strings.TrimSpace(string(out))
		}
	}
	if gobin != "" {
		path := filepath.Join(gobin, binName)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// 3. Check explicitly in GOPATH/bin
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		cmd := exec.Command("go", "env", "GOPATH")
		if out, err := cmd.Output(); err == nil {
			gopath = strings.TrimSpace(string(out))
		}
	}
	if gopath != "" {
		paths := strings.Split(gopath, string(os.PathListSeparator))
		for _, p := range paths {
			path := filepath.Join(p, "bin", binName)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	// 4. Try common default: ~/go/bin
	home, _ := os.UserHomeDir()
	if home != "" {
		path := filepath.Join(home, "go", "bin", binName)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// 5. Install if not found
	fmt.Printf("⚠️  %s not found in PATH or Go bin directories. Attempting to install...\n", binName)
	installCmd := exec.Command("go", "install", installURL)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to install %s: %w. Please install it manually with 'go install %s'", binName, err, installURL)
	}

	// After install, check again in common locations
	if path, err := exec.LookPath(binName); err == nil {
		return path, nil
	}
	
	// If still not in path, return the installation candidate if it exists now
	if gobin != "" {
		path := filepath.Join(gobin, binName)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	if home != "" {
		path := filepath.Join(home, "go", "bin", binName)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return binName, nil // Fallback to just name and hope for the best
}

func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("go.mod not found")
		}
		wd = parent
	}
}
