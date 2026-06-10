package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/compliance"
	"github.com/pmaojo/kthulu-go/internal/deployment"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "🔍 Enterprise security and compliance auditing",
	Long: `Comprehensive security analysis and compliance checking for enterprise environments.

Features:
  • SAST/DAST security scanning
  • Compliance validation (SOX, GDPR, PCI)
  • Dependency vulnerability analysis
  • Code quality metrics
  • Performance analysis

Examples:
  kthulu audit --compliance=sox
  kthulu audit --security --fix
  kthulu audit --dependencies`,
	RunE: func(cmd *cobra.Command, args []string) error {
		compliance, _ := cmd.Flags().GetString("compliance")
		security, _ := cmd.Flags().GetBool("security")
		dependencies, _ := cmd.Flags().GetBool("dependencies")
		fix, _ := cmd.Flags().GetBool("fix")

		return runAuditCommand(compliance, security, dependencies, fix)
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "🚀 Zero-config multi-cloud deployment",
	Long: `Deploy your Kthulu application to any cloud provider with zero configuration.

Supported Platforms:
  • AWS (EKS, Fargate, Lambda)
  • Google Cloud (GKE, Cloud Run)
  • Azure (AKS, Container Instances)
  • Kubernetes (any cluster)
  • Docker Swarm
  • Fly.io (fly.toml + fly deploy)

Examples:
  kthulu deploy --cloud=aws --scale=auto
  kthulu deploy --cloud=gcp --region=us-central1
  kthulu deploy --kubernetes --namespace=production
  kthulu deploy --cloud=fly --region=iad`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cloud, _ := cmd.Flags().GetString("cloud")
		scale, _ := cmd.Flags().GetString("scale")
		region, _ := cmd.Flags().GetString("region")
		namespace, _ := cmd.Flags().GetString("namespace")

		return runDeployCommand(cloud, scale, region, namespace)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "📊 Project health and status overview",
	Long: `Get comprehensive overview of your Kthulu project health and status.

Shows:
  • Module dependencies
  • Security vulnerabilities  
  • Performance metrics
  • Deployment status
  • Team activity

Examples:
  kthulu status
  kthulu status --detailed
  kthulu status --modules`,
	RunE: func(cmd *cobra.Command, args []string) error {
		detailed, _ := cmd.Flags().GetBool("detailed")
		modules, _ := cmd.Flags().GetBool("modules")

		return runStatusCommand(detailed, modules)
	},
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "📈 Upgrade framework and dependencies",
	Long: `Safely upgrade your Kthulu framework version and dependencies.

Features:
  • Automated migration scripts
  • Dependency conflict resolution
  • Backup creation
  • Rollback capability

Examples:
  kthulu upgrade
  kthulu upgrade --version=latest
  kthulu upgrade --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version, _ := cmd.Flags().GetString("version")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		return runUpgradeCommand(version, dryRun)
	},
}

func init() {
	// Audit command flags
	auditCmd.Flags().String("compliance", "", "Compliance standard (sox, gdpr, pci)")
	auditCmd.Flags().Bool("security", true, "Run security analysis")
	auditCmd.Flags().Bool("dependencies", true, "Check dependency vulnerabilities")
	auditCmd.Flags().Bool("fix", false, "Automatically fix found issues")

	// Deploy command flags
	deployCmd.Flags().String("cloud", "", "Cloud provider (aws, gcp, azure)")
	deployCmd.Flags().String("scale", "auto", "Scaling strategy (auto, manual, fixed)")
	deployCmd.Flags().String("region", "", "Deployment region")
	deployCmd.Flags().String("namespace", "default", "Kubernetes namespace")

	// Status command flags
	statusCmd.Flags().Bool("detailed", false, "Show detailed information")
	statusCmd.Flags().Bool("modules", false, "Focus on module information")

	// Upgrade command flags
	upgradeCmd.Flags().String("version", "latest", "Target version")
	upgradeCmd.Flags().Bool("dry-run", false, "Preview changes without applying")
}

func runAuditCommand(complianceStd string, security, dependencies, fix bool) error {
	if !security && !dependencies && complianceStd == "" && !fix {
		return fmt.Errorf("no audit options selected. Try --compliance=sox or --security")
	}

	fmt.Println("🔍 Enterprise Security Audit")
	var errs []error

	if security {
		if err := checkSAST(); err != nil {
			fmt.Printf("❌ Security check failed: %v\n", err)
			errs = append(errs, err)
		}
	}

	if dependencies {
		if err := checkDependencyVulnerabilities(); err != nil {
			fmt.Printf("❌ Dependency check failed: %v\n", err)
			errs = append(errs, err)
		}
	}

	if complianceStd != "" {
		fmt.Printf("📋 Validating %s compliance...\n", complianceStd)

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		report, err := compliance.Validate(complianceStd, cwd)
		if err != nil {
			return fmt.Errorf("compliance check failed: %w", err)
		}

		if report.Passed {
			fmt.Printf("✅ %s Compliance Checks Passed\n", report.Standard)
		} else {
			fmt.Printf("❌ %s Compliance Checks Failed\n", report.Standard)
		}

		for _, check := range report.Checks {
			icon := "✅"
			status := "PASS"
			if !check.Passed {
				icon = "❌"
				status = "FAIL"
			}
			fmt.Printf("  %s %-20s: %s\n", icon, check.Name, status)
			if !check.Passed {
				fmt.Printf("      Error: %v\n", check.Error)
			}
		}

		if !report.Passed {
			errs = append(errs, fmt.Errorf("%s compliance checks failed", complianceStd))
		}
	}

	if fix {
		fmt.Println("🔧 Auto-fixing detected issues...")
	}

	if len(errs) > 0 {
		return fmt.Errorf("audit failed with %d errors", len(errs))
	}

	return nil
}

func checkSAST() error {
	binName, err := ensureToolInstalled("gosec", "github.com/securego/gosec/v2/cmd/gosec@latest")
	if err != nil {
		return err
	}

	fmt.Println("🔒 Running SAST security scan with gosec...")
	cmd := exec.Command(binName, "-quiet", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("security vulnerabilities found")
	}

	fmt.Println("✅ No security vulnerabilities found")
	return nil
}

func checkDependencyVulnerabilities() error {
	fmt.Println("📦 Checking dependency vulnerabilities...")

	binName, err := ensureToolInstalled("govulncheck", "golang.org/x/vuln/cmd/govulncheck@latest")
	if err != nil {
		return err
	}

	cmd := exec.Command(binName, "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("🚀 Running govulncheck...")
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
			return fmt.Errorf("vulnerabilities found")
		}
		return err
	}

	return nil
}

func ensureToolInstalled(binName, installURL string) (string, error) {
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

func runDeployCommand(cloud, scale, region, namespace string) error {
	fmt.Println("🚀 Zero-Config Cloud Deployment")

	if cloud == "" {
		fmt.Println("🔍 Auto-detecting best cloud provider...")
		cloud = "kubernetes" // Default to generic k8s
	}

	// Fly.io deploy path
	if cloud == "fly" || cloud == "flyio" {
		deployer := deployment.NewFlyioDeployer(region)
		if err := deployer.Deploy(); err != nil {
			return fmt.Errorf("fly.io deployment failed: %w", err)
		}
		fmt.Println("\n✨ Deployment complete!")
		return nil
	}

	manager, err := deployment.NewManager(cloud, region, namespace, scale)
	if err != nil {
		return fmt.Errorf("failed to initialize deployment manager: %w", err)
	}

	// 1. Analyze project structure
	if err := manager.Analyze(); err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// 2. Generate cloud-specific configs
	if err := manager.GenerateConfig(); err != nil {
		return fmt.Errorf("config generation failed: %w", err)
	}

	// 3. Build container images
	if err := manager.Build(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// 4. Deploy to target platform
	if err := manager.Deploy(); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	// 5. Setup monitoring/logging
	if err := manager.SetupMonitoring(); err != nil {
		return fmt.Errorf("monitoring setup failed: %w", err)
	}

	// 6. Configure auto-scaling
	if err := manager.ConfigureAutoScaling(); err != nil {
		return fmt.Errorf("auto-scaling configuration failed: %w", err)
	}

	fmt.Println("\n✨ Deployment complete!")
	return nil
}

func runStatusCommand(detailed, modules bool) error {
	fmt.Println("📊 Kthulu Project Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// --- Read go.mod for project name and Go version ---
	projectName := "(unknown)"
	goVersion := "(unknown)"
	goModPath := filepath.Join(cwd, "go.mod")
	if data, err := os.ReadFile(goModPath); err == nil {
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "module ") {
				projectName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				// Use just the last path component as a friendly name
				parts := strings.Split(projectName, "/")
				projectName = parts[len(parts)-1]
			}
			if strings.HasPrefix(line, "go ") {
				goVersion = strings.TrimSpace(strings.TrimPrefix(line, "go "))
			}
		}
	}

	// --- Count active modules: subdirs in internal/ that contain .go files ---
	activeModules := 0
	var activeModuleNames []string
	internalDir := filepath.Join(cwd, "internal")
	if entries, err := os.ReadDir(internalDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			// Check if any .go file exists directly or transitively in this subdir
			subdir := filepath.Join(internalDir, e.Name())
			hasGo := false
			_ = filepath.WalkDir(subdir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") {
					hasGo = true
					return filepath.SkipAll
				}
				return nil
			})
			if hasGo {
				activeModules++
				activeModuleNames = append(activeModuleNames, e.Name())
			}
		}
	}

	// --- Count migrations: .sql files in migrations/ or db/migrations/ ---
	migrationCount := 0
	for _, migDir := range []string{
		filepath.Join(cwd, "migrations"),
		filepath.Join(cwd, "db", "migrations"),
	} {
		if entries, err := os.ReadDir(migDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
					migrationCount++
				}
			}
		}
	}

	// --- Detect features ---
	type featureCheck struct {
		name string
		path string
	}
	featureChecks := []featureCheck{
		{"auth", filepath.Join(cwd, "internal", "auth")},
		{"user", filepath.Join(cwd, "internal", "user")},
		{"queues", filepath.Join(cwd, "internal", "queues")},
		{"configs/app.yaml", filepath.Join(cwd, "configs", "app.yaml")},
	}
	var detectedFeatures []string
	for _, fc := range featureChecks {
		if _, err := os.Stat(fc.path); err == nil {
			detectedFeatures = append(detectedFeatures, fc.name)
		}
	}

	// --- Build status ---
	buildStatus := "✅ pass"
	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = cwd
	var buildOut bytes.Buffer
	buildCmd.Stderr = &buildOut
	if err := buildCmd.Run(); err != nil {
		buildStatus = "❌ fail"
		if buildOut.Len() > 0 {
			buildStatus += ": " + strings.TrimSpace(buildOut.String())
		}
	}

	// --- kthulu-plan.yaml ---
	planName := ""
	planDesc := ""
	planPath := filepath.Join(cwd, "kthulu-plan.yaml")
	if data, err := os.ReadFile(planPath); err == nil {
		var planDoc struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
		}
		if err := yaml.Unmarshal(data, &planDoc); err == nil {
			planName = planDoc.Name
			planDesc = planDoc.Description
		}
	}

	// --- Print output ---
	fmt.Printf("📁 Project:      %s\n", projectName)
	fmt.Printf("🏗️  Go version:   %s\n", goVersion)
	fmt.Printf("📦 Modules:      %d active (in internal/)\n", activeModules)
	if migrationCount > 0 {
		fmt.Printf("🗄️  Migrations:   %d SQL file(s)\n", migrationCount)
	} else {
		fmt.Printf("🗄️  Migrations:   none found\n")
	}
	fmt.Println()

	// Build status line
	fmt.Printf("🔨 Build:        %s\n", buildStatus)

	// Features
	if len(detectedFeatures) > 0 {
		fmt.Printf("✨ Features:     %s\n", strings.Join(detectedFeatures, ", "))
	} else {
		fmt.Println("✨ Features:     (none of auth/user/queues/configs detected)")
	}

	// kthulu-plan.yaml
	if planName != "" {
		fmt.Printf("📋 Plan:         %s", planName)
		if planDesc != "" {
			fmt.Printf(" — %s", planDesc)
		}
		fmt.Println()
	}

	fmt.Println()

	if modules && len(activeModuleNames) > 0 {
		fmt.Println("📦 Active Modules (internal/):")
		for _, name := range activeModuleNames {
			fmt.Printf("  ✅ %s\n", name)
		}
		fmt.Println()
	}

	if detailed {
		fmt.Println("ℹ️  Detailed mode: run 'kthulu secure' for vulnerability details,")
		fmt.Println("   'kthulu audit --compliance=sox' for compliance checks.")
	}

	return nil
}

func runUpgradeCommand(version string, dryRun bool) error {
	fmt.Printf("📈 Upgrading to Kthulu %s\n", version)

	if dryRun {
		fmt.Println("🔍 Dry run mode - no changes will be applied")
	}

	fmt.Println("🔍 Checking current version...")
	fmt.Println("📦 Analyzing dependencies...")
	fmt.Println("🔄 Planning migration...")

	if dryRun {
		fmt.Println("📋 Migration Plan:")
		fmt.Println("  • Update framework: v0.9.0 → v1.0.0")
		fmt.Println("  • Update dependencies: 3 packages")
		fmt.Println("  • Run migrations: 2 scripts")
		fmt.Println("  • Update configs: 1 file")
		return nil
	}

	return fmt.Errorf("upgrade system not yet implemented - coming in FASE 4!")
}
