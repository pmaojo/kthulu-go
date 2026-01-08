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
		diagrams, _ := cmd.Flags().GetBool("diagrams")
		return runDocCommand(cmd, args, diagrams)
	},
}

var (
	docDir         string
	docGeneralInfo string
)

func init() {
	docCmd.Flags().StringVar(&docDir, "dir", "", "Directory to search for main.go (default: project root)")
	docCmd.Flags().StringVar(&docGeneralInfo, "generalInfo", "", "API general info file (default: cmd/server/main.go)")
	docCmd.Flags().Bool("diagrams", false, "Generate Mermaid architecture diagrams")
	rootCmd.AddCommand(docCmd)
}

func runDocCommand(cmd *cobra.Command, args []string, generateDiagrams bool) error {
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

	// 3. Run Swagger Generation
	swagArgs := []string{"init"}
	if mainFile != "." {
		swagArgs = append(swagArgs, "--generalInfo", mainFile)
	}
	if projectRoot != "" {
		swagArgs = append(swagArgs, "--dir", projectRoot)
	}
	swagArgs = append(swagArgs, "--parseDependency", "--parseInternal")

	runCmd := exec.Command(binName, swagArgs...)
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

	// 4. Run Mermaid Generation (The WOW Feature)
	if generateDiagrams {
		fmt.Println("\n🎨 Generating Architecture Diagrams...")
		if err := generateMermaidDiagrams(projectRoot); err != nil {
			fmt.Printf("⚠️  Failed to generate diagrams: %v\n", err)
		}
	}

	fmt.Println("\n✅ Documentation generated successfully!")
	return nil
}

// generateMermaidDiagrams parses Go files and generates a class diagram
func generateMermaidDiagrams(root string) error {
	mermaidPath := filepath.Join(root, "docs", "architecture.mermaid")
	_ = os.MkdirAll(filepath.Join(root, "docs"), 0755)

	var sb strings.Builder
	sb.WriteString("classDiagram\n")
	sb.WriteString("    direction TB\n\n")

	structs := make(map[string]string) // name -> package
	relationships := []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if !strings.HasSuffix(path, ".go") { return nil }
		if strings.Contains(path, "test") || strings.Contains(path, "vendor") || strings.Contains(path, "mocks") { return nil }

		content, err := os.ReadFile(path)
		if err != nil { return nil }
		
		lines := strings.Split(string(content), "\n")
		packageName := "unknown"
		
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "package ") {
				packageName = strings.TrimPrefix(line, "package ")
				break
			}
		}

		for i, line := range lines {
			line = strings.TrimSpace(line)
			// Detect struct definitions
			if strings.HasPrefix(line, "type ") && strings.Contains(line, " struct {") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					structName := parts[1]
					structs[structName] = packageName
					
					// Naive relationship detection in fields
					for j := i + 1; j < len(lines); j++ {
						fieldLine := strings.TrimSpace(lines[j])
						if fieldLine == "}" { break }
						if fieldLine == "" || strings.HasPrefix(fieldLine, "//") { continue }
						
						fParts := strings.Fields(fieldLine)
						if len(fParts) >= 2 {
							fieldType := fParts[1]
							// Look for other structs being mentioned (ignoring pointers, slices for now)
							cleanType := strings.TrimLeft(fieldType, "*[]")
							if cleanType != "" && cleanType[0] >= 'A' && cleanType[0] <= 'Z' {
								relationships = append(relationships, fmt.Sprintf("    %s --> %s", structName, cleanType))
							}
						}
					}
				}
			}
			
			// Detect interface definitions
			if strings.HasPrefix(line, "type ") && strings.Contains(line, " interface {") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					ifName := parts[1]
					sb.WriteString(fmt.Sprintf("    class %s {\n", ifName))
					sb.WriteString("        <<interface>>\n")
					sb.WriteString(fmt.Sprintf("        %s\n", packageName))
					sb.WriteString("    }\n")
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Output classes
	for name, pkg := range structs {
		sb.WriteString(fmt.Sprintf("    class %s {\n", name))
		sb.WriteString(fmt.Sprintf("        %s\n", pkg))
		sb.WriteString("    }\n")
	}

	// Output unique relationships
	relMap := make(map[string]bool)
	for _, rel := range relationships {
		if !relMap[rel] {
			sb.WriteString(rel + "\n")
			relMap[rel] = true
		}
	}

	if err := os.WriteFile(mermaidPath, []byte(sb.String()), 0644); err != nil {
		return err
	}
	
	fmt.Printf("   Generated: %s\n", mermaidPath)
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
