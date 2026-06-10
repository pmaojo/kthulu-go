package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var secureCmd = &cobra.Command{
	Use:   "secure",
	Short: "🔐 Run security checks: vet, gosec, govulncheck",
	Long: `Run a suite of security and correctness checks on your Go project.

Checks:
  • go vet ./...         — static analysis for correctness issues
  • gosec ./...          — SAST security scanner (installed if missing)
  • govulncheck ./...    — dependency vulnerability database check (installed if missing)

Produces a formatted summary with issue counts by severity and returns a
non-zero exit code when HIGH severity issues are found.

Examples:
  kthulu secure
  kthulu secure --fix`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fix, _ := cmd.Flags().GetBool("fix")
		return runSecureCommand(fix)
	},
}

func init() {
	secureCmd.Flags().Bool("fix", false, "Run go fix ./... before security checks")
	rootCmd.AddCommand(secureCmd)
}

// secureResult holds the outcome of a single security tool run.
type secureResult struct {
	vetErrors  int
	highCount  int
	medCount   int
	lowCount   int
	vulnCount  int
	vetOutput  string
	gosecOutput string
	vulnOutput string
}

func runSecureCommand(fix bool) error {
	fmt.Println("🔐 Kthulu Secure — Security & Correctness Scan")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	result := &secureResult{}

	// --fix: run go fix ./... first
	if fix {
		fmt.Println("🔧 Running go fix ./...")
		fixCmd := exec.Command("go", "fix", "./...")
		fixCmd.Stdout = os.Stdout
		fixCmd.Stderr = os.Stderr
		if err := fixCmd.Run(); err != nil {
			fmt.Printf("⚠️  go fix returned an error: %v\n", err)
		} else {
			fmt.Println("✅ go fix completed")
		}
		fmt.Println()
	}

	// Step 1: go vet
	if err := runGoVetSecure(result); err != nil {
		fmt.Printf("⚠️  go vet encountered an error: %v\n", err)
	}

	// Step 2: gosec
	if err := runGosecSecure(result); err != nil {
		fmt.Printf("⚠️  gosec encountered an error: %v\n", err)
	}

	// Step 3: govulncheck
	if err := runGovulncheckSecure(result); err != nil {
		fmt.Printf("⚠️  govulncheck encountered an error: %v\n", err)
	}

	// Print summary
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📋 Security Scan Summary")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	vetIcon := "✅"
	if result.vetErrors > 0 {
		vetIcon = "❌"
	}
	fmt.Printf("  %s go vet errors:        %d\n", vetIcon, result.vetErrors)

	highIcon := "✅"
	if result.highCount > 0 {
		highIcon = "🔴"
	}
	fmt.Printf("  %s gosec HIGH issues:    %d\n", highIcon, result.highCount)

	medIcon := "✅"
	if result.medCount > 0 {
		medIcon = "🟡"
	}
	fmt.Printf("  %s gosec MEDIUM issues:  %d\n", medIcon, result.medCount)

	lowIcon := "✅"
	if result.lowCount > 0 {
		lowIcon = "🔵"
	}
	fmt.Printf("  %s gosec LOW issues:     %d\n", lowIcon, result.lowCount)

	vulnIcon := "✅"
	if result.vulnCount > 0 {
		vulnIcon = "❌"
	}
	fmt.Printf("  %s vulnerabilities:      %d\n", vulnIcon, result.vulnCount)

	fmt.Println()

	if result.highCount > 0 {
		fmt.Printf("❌ FAILED: %d HIGH severity issue(s) found — resolve before shipping.\n", result.highCount)
		return fmt.Errorf("security scan failed: %d HIGH severity issue(s)", result.highCount)
	}

	total := result.vetErrors + result.medCount + result.lowCount + result.vulnCount
	if total > 0 {
		fmt.Println("⚠️  Scan completed with warnings. No HIGH severity issues.")
	} else {
		fmt.Println("🎉 All security checks passed — no issues found!")
	}
	return nil
}

// runGoVetSecure runs go vet ./... and captures output.
func runGoVetSecure(result *secureResult) error {
	fmt.Println("🔍 Running go vet ./...")
	var buf bytes.Buffer
	cmd := exec.Command("go", "vet", "./...")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	output := buf.String()
	result.vetOutput = output

	if output != "" {
		fmt.Print(output)
		// Count non-empty lines as errors
		for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
			if strings.TrimSpace(line) != "" {
				result.vetErrors++
			}
		}
	}
	if err == nil {
		fmt.Println("✅ go vet: no issues found")
	}
	fmt.Println()
	return nil
}

// runGosecSecure installs gosec if needed and runs it, parsing severity counts.
func runGosecSecure(result *secureResult) error {
	fmt.Println("🔒 Running gosec -fmt=text -quiet ./...")
	binName, err := ensureToolInstalled("gosec", "github.com/securego/gosec/v2/cmd/gosec@latest")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	cmd := exec.Command(binName, "-fmt=text", "-quiet", "./...")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	// gosec returns non-zero when it finds issues; we handle that manually
	_ = cmd.Run()
	output := buf.String()
	result.gosecOutput = output

	if strings.TrimSpace(output) != "" {
		fmt.Print(output)
	}

	// Parse severity counts from gosec text output.
	// gosec text format includes lines like "[HIGH]", "[MEDIUM]", "[LOW]"
	for _, line := range strings.Split(output, "\n") {
		upper := strings.ToUpper(line)
		switch {
		case strings.Contains(upper, "[HIGH]"):
			result.highCount++
		case strings.Contains(upper, "[MEDIUM]"):
			result.medCount++
		case strings.Contains(upper, "[LOW]"):
			result.lowCount++
		}
	}

	if result.highCount == 0 && result.medCount == 0 && result.lowCount == 0 {
		fmt.Println("✅ gosec: no issues found")
	}
	fmt.Println()
	return nil
}

// runGovulncheckSecure installs govulncheck if needed and runs it.
func runGovulncheckSecure(result *secureResult) error {
	fmt.Println("📦 Running govulncheck ./...")
	binName, err := ensureToolInstalled("govulncheck", "golang.org/x/vuln/cmd/govulncheck@latest")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	cmd := exec.Command(binName, "./...")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	runErr := cmd.Run()
	output := buf.String()
	result.vulnOutput = output

	if strings.TrimSpace(output) != "" {
		fmt.Print(output)
	}

	// govulncheck exits with code 3 when vulnerabilities are found.
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
			// Count "Vulnerability #N:" lines
			for _, line := range strings.Split(output, "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "Vulnerability #") {
					result.vulnCount++
				}
			}
			if result.vulnCount == 0 {
				result.vulnCount = 1 // at least one if exit code 3
			}
		}
	}

	if result.vulnCount == 0 {
		fmt.Println("✅ govulncheck: no vulnerabilities found")
	}
	fmt.Println()
	return nil
}
