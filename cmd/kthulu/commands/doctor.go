package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "🩺 Diagnose your development environment",
	Long: `Diagnose your development environment and Kthulu project configuration.

Checks for:
  • Essential tools (Go, Git, Docker, Make)
  • Frontend tools (Bun/Node)
  • Project configuration (if running inside a Kthulu project)
  • Environment variables

Usage:
  kthulu doctor
  kthulu doctor --verbose`,
	Run: runDoctor,
}

func init() {
	// Flags
	doctorCmd.Flags().BoolP("verbose", "v", false, "Show detailed output for checks")
}

type diagnosticCheck struct {
	Category    string
	Name        string
	Run         func() (string, error) // Returns version/info string, or error
	Remedy      string
	Essential   bool
}

func runDoctor(cmd *cobra.Command, args []string) {
	verbose, _ := cmd.Flags().GetBool("verbose")

	fmt.Println("🩺 Kthulu Doctor")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	checks := []diagnosticCheck{
		// Core Tools
		{
			Category:  "Core Tools",
			Name:      "Go",
			Essential: true,
			Run: func() (string, error) {
				out, err := exec.Command("go", "version").Output()
				if err != nil {
					return "", err
				}
				return strings.TrimSpace(string(out)), nil
			},
			Remedy: "Install Go from https://go.dev/dl/",
		},
		{
			Category:  "Core Tools",
			Name:      "Git",
			Essential: true,
			Run: func() (string, error) {
				out, err := exec.Command("git", "--version").Output()
				if err != nil {
					return "", err
				}
				return strings.TrimSpace(string(out)), nil
			},
			Remedy: "Install Git from https://git-scm.com/downloads",
		},
		{
			Category:  "Core Tools",
			Name:      "Make",
			Essential: false,
			Run: func() (string, error) {
				out, err := exec.Command("make", "--version").Output()
				if err != nil {
					return "", err
				}
				// Make version output is multi-line, take first line
				lines := strings.Split(string(out), "\n")
				return strings.TrimSpace(lines[0]), nil
			},
			Remedy: "Install Make (typically part of build-essential or Xcode Command Line Tools)",
		},
		{
			Category:  "Infrastructure",
			Name:      "Docker",
			Essential: false,
			Run: func() (string, error) {
				// Just check if docker is in path first
				path, err := exec.LookPath("docker")
				if err != nil {
					return "", fmt.Errorf("docker not found in PATH")
				}
				// Check if daemon is running
				out, err := exec.Command("docker", "info").CombinedOutput()
				if err != nil {
					// Check if it's a permission issue or daemon issue
					if strings.Contains(string(out), "permission denied") {
						return path, fmt.Errorf("docker installed but permission denied (check user group)")
					}
					if strings.Contains(string(out), "Is the docker daemon running") {
						return path, fmt.Errorf("docker installed but daemon not running")
					}
					return "", fmt.Errorf("docker check failed: %v", err)
				}
				return fmt.Sprintf("Running at %s", path), nil
			},
			Remedy: "Install Docker Desktop or ensure the Docker daemon is running",
		},
		// Frontend
		{
			Category:  "Frontend",
			Name:      "Bun",
			Essential: false, // Recommended but not strictly essential if using Node, though Kthulu prefers Bun
			Run: func() (string, error) {
				out, err := exec.Command("bun", "--version").Output()
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("bun %s", strings.TrimSpace(string(out))), nil
			},
			Remedy: "Install Bun from https://bun.sh (curl -fsSL https://bun.sh/install | bash)",
		},
		{
			Category:  "Frontend",
			Name:      "Node.js",
			Essential: false,
			Run: func() (string, error) {
				out, err := exec.Command("node", "--version").Output()
				if err != nil {
					return "", err
				}
				return strings.TrimSpace(string(out)), nil
			},
			Remedy: "Install Node.js from https://nodejs.org (LTS recommended)",
		},
	}

	// Project Checks (if applicable)
	wd, _ := os.Getwd()
	inProject := isKthuluProject(wd)

	if inProject {
		checks = append(checks, diagnosticCheck{
			Category:  "Project",
			Name:      "Kthulu Config",
			Essential: true,
			Run: func() (string, error) {
				// Check for go.mod
				if _, err := os.Stat(filepath.Join(wd, "go.mod")); err != nil {
					return "", fmt.Errorf("missing go.mod")
				}
				// Check for cmd/server
				if _, err := os.Stat(filepath.Join(wd, "cmd", "server", "main.go")); err != nil {
					return "", fmt.Errorf("missing cmd/server/main.go")
				}
				return fmt.Sprintf("Valid project at %s", wd), nil
			},
			Remedy: "Run 'kthulu create' to generate a valid project structure",
		})
	}

	// Execution
	currentCategory := ""
	hasFailures := false
	warnings := []string{}

	for _, check := range checks {
		if check.Category != currentCategory {
			fmt.Printf("\n[%s]\n", check.Category)
			currentCategory = check.Category
		}

		info, err := check.Run()
		if err == nil {
			fmt.Printf("  ✅  %-15s %s\n", check.Name, info)
		} else {
			icon := "⚠️ "
			if check.Essential {
				icon = "❌ "
				hasFailures = true
			} else {
				warnings = append(warnings, fmt.Sprintf("%s: %s", check.Name, check.Remedy))
			}

			fmt.Printf("  %s %-15s %v\n", icon, check.Name, err)

			if verbose {
				fmt.Printf("      Remedy: %s\n", check.Remedy)
			}
		}
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if hasFailures {
		fmt.Println("❌ Doctor found issues that need attention.")
		if !verbose {
			fmt.Println("   Run with --verbose for detailed fix instructions.")
		}
		os.Exit(1)
	} else if len(warnings) > 0 {
		fmt.Println("⚠️  Doctor passed with warnings.")
		if verbose {
			for _, w := range warnings {
				fmt.Printf("   • %s\n", w)
			}
		}
	} else {
		fmt.Println("🎉 Everything looks good! You are ready to build.")
	}
}
