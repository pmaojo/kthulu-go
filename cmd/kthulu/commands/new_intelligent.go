package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/internal/blueprint"
	"github.com/pmaojo/kthulu-go/internal/generator"
	"github.com/pmaojo/kthulu-go/internal/resolver"
)

// Template definitions
type ProjectTemplate struct {
	Name        string
	Description string
	Features    []string
	Database    string
	Frontend    string
	Auth        string
	Enterprise  bool
}

var projectTemplates = map[string]ProjectTemplate{
	"microservice": {
		Name:        "Microservice",
		Description: "Lightweight microservice with essential features",
		Features:    []string{"user", "auth"},
		Database:    "sqlite",
		Frontend:    "templ",
		Auth:        "jwt",
		Enterprise:  false,
	},
	"monolith": {
		Name:        "Monolith",
		Description: "Full-featured monolithic application",
		Features:    []string{"user", "auth", "organization", "contact", "product"},
		Database:    "postgres",
		Frontend:    "templ",
		Auth:        "jwt",
		Enterprise:  false,
	},
	"api-gateway": {
		Name:        "API Gateway",
		Description: "API Gateway with routing and load balancing",
		Features:    []string{"user", "auth", "oauthsso", "realtime"},
		Database:    "postgres",
		Frontend:    "templ",
		Auth:        "oauth",
		Enterprise:  true,
	},
	"fintech": {
		Name:        "FinTech",
		Description: "Financial services with compliance and security",
		Features:    []string{"user", "auth", "organization", "contact", "product", "invoice", "payment", "verifactu", "audit"},
		Database:    "postgres",
		Frontend:    "templ",
		Auth:        "both",
		Enterprise:  true,
	},
	"ecommerce": {
		Name:        "E-commerce",
		Description: "Complete e-commerce platform",
		Features:    []string{"user", "auth", "organization", "contact", "product", "inventory", "invoice", "payment", "notification", "calendar"},
		Database:    "postgres",
		Frontend:    "templ",
		Auth:        "oauth",
		Enterprise:  false,
	},
	"saas": {
		Name:        "SaaS Platform",
		Description: "Multi-tenant SaaS application",
		Features:    []string{"user", "auth", "organization", "contact", "product", "invoice", "payment", "oauthsso", "notification", "audit", "realtime"},
		Database:    "postgres",
		Frontend:    "templ",
		Auth:        "both",
		Enterprise:  true,
	},
	"cli": {
		Name:        "CLI Tool",
		Description: "Command Line Interface application using Cobra",
		Features:    []string{},
		Database:    "none",
		Frontend:    "none",
		Auth:        "none",
		Enterprise:  false,
	},
	"mcp": {
		Name:        "MCP Server",
		Description: "Model Context Protocol server for AI agents",
		Features:    []string{},
		Database:    "none",
		Frontend:    "none",
		Auth:        "none",
		Enterprise:  false,
	},
}

var newCmd = &cobra.Command{
	Use:     "create [name]",
	Aliases: []string{"new"},
	Short:   "🚀 Create a new Kthulu project (The easy way)",
	Long: `Build a production-ready application in seconds.

This command sets up everything you need:
  • Backend (Go + Chi + GORM)
  • Frontend (React/Next.js or similar)
  • Database (SQLite/Postgres)
  • Authentication (JWT/OAuth)

How to use:
  1. Simple start:
     kthulu new my-app

  2. With specific features:
     kthulu new my-shop --template=ecommerce --from-plan=plan.yaml

  3. Full stack (recommended):
     kthulu new airbnb-clone --frontend=react --features=user,auth,property

The tool will automatically resolve dependencies, create the folder structure,
and prepare your development environment.

IMPORTANT — entity fields: features without declared fields default to a
single 'name' column. For real applications declare every entity's fields
through a plan file (modules.<name>.fields with name:type[:rules] syntax)
and pass --from-plan, or use the MCP scaffold_project tool which takes the
domain model as structured data.
`,
	Args: cobra.ExactArgs(1),
	Run:  runNewProjectIntelligent,
}

var (
	newTemplate        string
	newFeatures        []string
	newDatabase        string
	newFrontend        string
	newAuth            string
	newModulePath      string
	newEnterprise      bool
	newObservability   bool
	newNoVercel        bool
	newOutputPath      string
	newDryRun          bool
	newSkipPostgen     bool
	newInteractive     bool
	newFromPlan        string
	newModuleFields    map[string][]string
	newFrontendModules []string
	newRequirements    []blueprint.Requirement
)

const (
	recommendedCoveragePercent = 60.0
)

func init() {
	newCmd.Flags().StringVarP(&newTemplate, "template", "t", "microservice", "Project template")
	newCmd.Flags().StringSliceVarP(&newFeatures, "features", "f", []string{}, "Comma-separated list of features/modules")
	newCmd.Flags().StringVarP(&newDatabase, "database", "d", "", "Database type (sqlite, postgres, mysql)")
	newCmd.Flags().StringVar(&newFrontend, "frontend", "templ", "Frontend type: templ (GTH - default), none")
	newCmd.Flags().StringVar(&newAuth, "auth", "", "Auth type (jwt, oauth, both)")
	newCmd.Flags().StringVar(&newModulePath, "module-path", "", "Go module path (default: project name)")
	newCmd.Flags().BoolVar(&newEnterprise, "enterprise", false, "Enable enterprise features")
	newCmd.Flags().BoolVar(&newObservability, "observability", false, "Enable observability stack")
	newCmd.Flags().BoolVar(&newNoVercel, "no-vercel", false, "Skip Vercel deployment files")
	newCmd.Flags().StringVarP(&newOutputPath, "output", "o", "", "Output directory (default: current directory)")
	newCmd.Flags().BoolVar(&newDryRun, "dry-run", false, "Show what would be generated without creating files")
	newCmd.Flags().BoolVar(&newSkipPostgen, "skip-postgen", false, "Skip templ generate, go mod tidy and go test after generation (fast mode; default in MCP)")
	newCmd.Flags().BoolVar(&newInteractive, "interactive", false, "Interactive project configuration")
	newCmd.Flags().StringVar(&newFromPlan, "from-plan", "", "Create project from plan file")

	rootCmd.AddCommand(newCmd)
}

func runNewProjectIntelligent(cmd *cobra.Command, args []string) {
	var projectName string
	var rawInput string

	if len(args) > 0 {
		rawInput = args[0]
		// Sanitize project name to be a valid module name (last element of path)
		projectName = filepath.Base(rawInput)

		// If input looks like a path and output path wasn't explicitly set, use the input as output path
		if (strings.Contains(rawInput, "/") || strings.Contains(rawInput, "\\")) && newOutputPath == "" {
			newOutputPath = rawInput
		}
	} else if newFromPlan != "" {
		// Will be extracted from plan
	} else {
		cmd.Help()
		os.Exit(1)
	}

	fmt.Printf("🧠 Creating intelligent Kthulu project: %s\n", projectName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Step 1: Initialize configuration
	config, err := buildProjectConfig(projectName)
	if err != nil {
		fmt.Printf("❌ Error building configuration: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Interactive mode if requested
	if newInteractive {
		if err := runInteractiveMode(config); err != nil {
			fmt.Printf("❌ Error in interactive mode: %v\n", err)
			os.Exit(1)
		}
	}

	// Step 3: Display configuration
	displayProjectConfiguration(config)

	// Step 4: Initialize intelligent system
	analysis := &parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
		Tags:         []parser.Tag{},
	}

	dependencyResolver := resolver.NewDependencyResolver(analysis)
	templateGenerator := generator.NewTemplateGenerator(dependencyResolver)

	// Step 5: Generate project structure
	fmt.Println("\n🏗️  Generating project structure...")
	structure, err := templateGenerator.GenerateProject(config)
	if err != nil {
		fmt.Printf("❌ Error generating project: %v\n", err)
		os.Exit(1)
	}

	// Step 6: Display generation plan
	displayGenerationPlan(structure)

	// Step 7: Write files (unless dry-run)
	if newDryRun {
		fmt.Println("\n🔍 Dry run completed - no files were created")
		return
	}

	fmt.Println("\n📁 Writing project files...")
	if err := templateGenerator.WriteProject(structure); err != nil {
		fmt.Printf("❌ Error writing project: %v\n", err)
		os.Exit(1)
	}

	// Step 7b: Write Requirements if present
	if len(newRequirements) > 0 {
		if err := writeRequirementsDoc(structure.RootPath, newRequirements); err != nil {
			fmt.Printf("⚠️  Warning: Failed to write REQUIREMENTS.md: %v\n", err)
		} else {
			fmt.Println("   📄 Generated docs/REQUIREMENTS.md")
		}
	}

	// Step 7c: Copy Plan file if used
	if newFromPlan != "" {
		if err := copyPlanFile(newFromPlan, structure.RootPath); err != nil {
			fmt.Printf("⚠️  Warning: Failed to copy plan file: %v\n", err)
		} else {
			fmt.Println("   📄 Copied plan file to project root")
		}
	}

	// Steps 8-9 (templ generate, go mod tidy, go test) download module
	// dependencies and can take minutes on a cold cache. Skip them when
	// requested or when driven through MCP, where client tool calls have
	// short timeouts — the agent can run the printed commands itself.
	if newSkipPostgen || os.Getenv("KTHULU_MCP_MODE") == "1" {
		fmt.Println("\n⏩ Skipping post-generation steps (templ generate, go mod tidy, go test).")
		fmt.Println("   Finish setup with:")
		fmt.Printf("     cd %s\n", structure.RootPath)
		if config.Frontend == "templ" {
			fmt.Println("     go run github.com/a-h/templ/cmd/templ@v0.3.977 generate ./...")
		}
		fmt.Println("     go mod tidy")
		fmt.Println("     go test ./...")
		displaySuccessMessage(projectName, config, structure)
		return
	}

	// Step 8: Run templ generate first so view packages contain Go files
	// before go mod tidy resolves imports.
	if config.Frontend == "templ" {
		if err := runTemplGenerate(structure.RootPath); err != nil {
			fmt.Printf("⚠️  Warning: Failed to run templ generate: %v\n", err)
		}
	}

	// Step 8b: Run go mod tidy
	if err := runGoModTidy(structure.RootPath); err != nil {
		fmt.Printf("❌ Error running go mod tidy: %v\n", err)
		os.Exit(1)
	}

	// Step 9: Execute tests with coverage requirements
	if err := runGoTests(structure.RootPath); err != nil {
		fmt.Printf("❌ Error running go test: %v\n", err)
		os.Exit(1)
	}

	// Step 10: Display success message and next steps
	displaySuccessMessage(projectName, config, structure)
}

func runGoTests(projectPath string) error {
	fmt.Println("\n🧪 Running go test with coverage...")
	testCmd := exec.Command("go", "test", "./...", "-coverprofile=coverage.out")
	testCmd.Dir = projectPath
	testCmd.Env = append(os.Environ(), "GOWORK=off") // Avoid parent go.work interference
	testCmd.Stdout = os.Stdout
	testCmd.Stderr = os.Stderr
	if err := testCmd.Run(); err != nil {
		fmt.Printf("⚠️  Tests failed (non-fatal): %v\n", err)
		fmt.Println("   This is often caused by go.work files. Run 'GOWORK=off go test ./...' manually.")
		return nil // Don't fail project creation on test failures
	}

	coverage, err := readCoveragePercentage(projectPath)
	if err != nil {
		fmt.Printf("⚠️  Could not read coverage: %v\n", err)
		return nil
	}

	_ = os.Remove(filepath.Join(projectPath, "coverage.out"))

	fmt.Printf("✅ Tests passed. Coverage: %.1f%%\n", coverage)
	if coverage < recommendedCoveragePercent {
		fmt.Printf("⚠️  Coverage below recommended %.1f%%. Consider adding more tests.\n", recommendedCoveragePercent)
	}

	return nil
}

func readCoveragePercentage(projectPath string) (float64, error) {
	coverageCmd := exec.Command("go", "tool", "cover", "-func=coverage.out")
	coverageCmd.Dir = projectPath
	var buffer bytes.Buffer
	coverageCmd.Stdout = &buffer
	coverageCmd.Stderr = os.Stderr
	if err := coverageCmd.Run(); err != nil {
		return 0, fmt.Errorf("coverage report failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	if len(lines) == 0 {
		return 0, fmt.Errorf("no coverage information produced")
	}

	lastLine := lines[len(lines)-1]
	fields := strings.Fields(lastLine)
	if len(fields) == 0 {
		return 0, fmt.Errorf("unexpected coverage output: %s", lastLine)
	}

	percentage := strings.TrimSuffix(fields[len(fields)-1], "%")
	coverage, err := strconv.ParseFloat(percentage, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse coverage: %w", err)
	}

	return coverage, nil
}

func buildProjectConfig(projectName string) (*generator.GeneratorConfig, error) {
	if newFromPlan != "" {
		planData, err := os.ReadFile(newFromPlan)
		if err != nil {
			return nil, fmt.Errorf("failed to read plan file: %w", err)
		}
		var bp blueprint.ProjectBlueprint
		if err := yaml.Unmarshal(planData, &bp); err != nil {
			return nil, fmt.Errorf("failed to parse plan file: %w", err)
		}
		fmt.Printf("📋 Loaded plan from %s (Template: %s)\n", newFromPlan, bp.Template)

		// Override config from plan
		if projectName == "" {
			projectName = bp.Name
		}
		newTemplate = bp.Template
		newFeatures = bp.Features
		newDatabase = bp.Database
		newFrontend = bp.Frontend
		newAuth = bp.Auth
		newRequirements = bp.Requirements

		// Populate parsed features and fields
		newFeatures = bp.Features // Start with base features list
		parsedModuleFields := make(map[string][]string)

		seenFeatures := make(map[string]bool, len(newFeatures))
		for _, f := range newFeatures {
			seenFeatures[f] = true
		}
		for name, config := range bp.Modules {
			if !seenFeatures[name] {
				newFeatures = append(newFeatures, name)
				seenFeatures[name] = true
			}
			newFrontendModules = append(newFrontendModules, name) // Modules get frontend
			if len(config.Fields) > 0 {
				parsedModuleFields[name] = config.Fields
			}
		}

		newModuleFields = parsedModuleFields
	}

	// Start with template defaults
	template, exists := projectTemplates[newTemplate]
	if !exists {
		return nil, fmt.Errorf("unknown template: %s", newTemplate)
	}

	config := &generator.GeneratorConfig{
		ProjectName:     projectName,
		OutputPath:      getOutputPath(projectName),
		Frontend:        template.Frontend,
		Database:        template.Database,
		Auth:            template.Auth,
		Features:        template.Features,
		Enterprise:      template.Enterprise,
		Observability:   false,
		ModuleFields:    newModuleFields,
		FrontendModules: newFrontendModules, // Convention: 'modules:' = fullstack
		CustomValues:    make(map[string]string),
	}

	if newTemplate == "cli" || newTemplate == "mcp" {
		config.TemplateType = newTemplate
	} else {
		config.TemplateType = "server"
	}

	if newModulePath != "" {
		config.CustomValues["module_path"] = newModulePath
	}

	if newNoVercel {
		config.CustomValues["no_vercel"] = "true"
	}

	// Override with command flags
	if len(newFeatures) > 0 {
		config.Features = newFeatures
	}
	if newDatabase != "" {
		config.Database = newDatabase
	}
	if newFrontend != "none" && newFrontend != "" {
		config.Frontend = newFrontend
	}
	if newAuth != "" {
		config.Auth = newAuth
	}
	if newEnterprise {
		config.Enterprise = true
	}
	if newObservability {
		config.Observability = true
	}

	return config, nil
}

func getOutputPath(projectName string) string {
	cwd, _ := os.Getwd()

	// If explicit output path is provided
	if newOutputPath != "" {
		if filepath.IsAbs(newOutputPath) {
			return newOutputPath
		}
		return filepath.Join(cwd, newOutputPath)
	}

	// Default: Create folder in current directory
	return filepath.Join(cwd, projectName)
}

func runInteractiveMode(config *generator.GeneratorConfig) error {
	// Interactive configuration would go here
	// For now, just return without changes
	fmt.Println("📝 Interactive mode not yet implemented - using current configuration")
	return nil
}

func displayProjectConfiguration(config *generator.GeneratorConfig) {
	fmt.Printf("\n📋 Project Configuration:\n")
	fmt.Printf("   Name:          %s\n", config.ProjectName)
	fmt.Printf("   Path:          %s\n", config.OutputPath)
	fmt.Printf("   Template:      %s\n", newTemplate)
	fmt.Printf("   Features:      %s\n", strings.Join(config.Features, ", "))
	fmt.Printf("   Database:      %s\n", config.Database)
	fmt.Printf("   Frontend:      %s\n", config.Frontend)
	fmt.Printf("   Auth:          %s\n", config.Auth)
	fmt.Printf("   Enterprise:    %v\n", config.Enterprise)
	fmt.Printf("   Observability: %v\n", config.Observability)
}

func displayGenerationPlan(structure *generator.ProjectStructure) {
	fmt.Printf("\n📊 Generation Plan:\n")
	fmt.Printf("   Directories:  %d\n", len(structure.Directories))
	fmt.Printf("   Files:        %d\n", len(structure.Files))
	fmt.Printf("   Dependencies: %d modules\n", len(structure.Dependencies))

	if len(structure.Dependencies) > 0 {
		fmt.Printf("   \nModules:      %s\n", strings.Join(structure.Dependencies, ", "))
	}

	// Show first few files that will be generated
	fmt.Printf("\n   Key files:\n")
	count := 0
	for _, file := range structure.Files {
		if count >= 5 {
			break
		}
		fmt.Printf("     • %s\n", file.Path)
		count++
	}
	if len(structure.Files) > 5 {
		fmt.Printf("     • ... and %d more files\n", len(structure.Files)-5)
	}
}

func writeRequirementsDoc(rootPath string, requirements []blueprint.Requirement) error {
	docsDir := filepath.Join(rootPath, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return err
	}

	var builder strings.Builder
	builder.WriteString("# Project Requirements\n\n")
	builder.WriteString("| ID | Title | Priority | Status |\n")
	builder.WriteString("|---|---|---|---|\n")

	for _, req := range requirements {
		builder.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", req.ID, req.Title, req.Priority, req.Status))
	}

	builder.WriteString("\n## Details\n\n")
	for _, req := range requirements {
		builder.WriteString(fmt.Sprintf("### %s: %s\n", req.ID, req.Title))
		if req.Description != "" {
			builder.WriteString(req.Description + "\n\n")
		}
		builder.WriteString(fmt.Sprintf("- **Priority:** %s\n", req.Priority))
		builder.WriteString(fmt.Sprintf("- **Status:** %s\n", req.Status))
		builder.WriteString(fmt.Sprintf("- **Created:** %s\n\n", req.Created))
	}

	return os.WriteFile(filepath.Join(docsDir, "REQUIREMENTS.md"), []byte(builder.String()), 0644)
}

func copyPlanFile(src, destRoot string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(destRoot, "kthulu-plan.yaml"), input, 0644)
}

func displaySuccessMessage(projectName string, config *generator.GeneratorConfig, structure *generator.ProjectStructure) {
	fmt.Printf("\n🎉 Project '%s' created successfully!\n", projectName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Printf("\n📍 Location: %s\n", structure.RootPath)
	fmt.Printf("📦 Modules:  %d (%s)\n", len(structure.Dependencies), strings.Join(structure.Dependencies, ", "))
	fmt.Printf("📁 Files:    %d generated\n", len(structure.Files))

	fmt.Printf("\n🚀 Next steps:\n")
	fmt.Printf("   cd %s\n", projectName)

	if config.Database != "sqlite" {
		fmt.Printf("   # Configure %s connection in configs/app.yaml\n", config.Database)
	}

	fmt.Printf("   kthulu migrate up           # Run database migrations\n")
	fmt.Printf("   kthulu dev                  # Start dev server with AI self-healing\n")

	fmt.Printf("\n💡 Troubleshooting:\n")
	fmt.Printf("   If you see 'local package' errors, try: export GOWORK=off\n")
	fmt.Printf("   Or add the project to your workspace: go work use .\n")

	if config.Frontend == "react" {
		fmt.Printf("\n💻 Frontend development:\n")
		fmt.Printf("   cd %s/frontend\n", projectName)
		fmt.Printf("   npm install\n")
		fmt.Printf("   npm run dev\n")
	}

	fmt.Printf("\n🔧 Additional commands:\n")
	fmt.Printf("   kthulu add module <name>    # Add new modules\n")
	fmt.Printf("   kthulu add component <type> # Add new components (handler, service, etc)\n")
	fmt.Printf("   kthulu ai suggest           # Get AI recommendations\n")
	fmt.Printf("   kthulu analyze              # Analyze project structure\n")
	fmt.Printf("   kthulu doc                  # Generate OpenAPI documentation\n")
	fmt.Printf("   kthulu secure               # Audit and patch security vulnerabilities\n")
	fmt.Printf("   kthulu audit                # Run enterprise compliance checks\n")
	fmt.Printf("   kthulu deploy               # Deploy to cloud providers\n")

	fmt.Printf("\n📚 Documentation: https://docs.kthulu.dev\n")
	fmt.Printf("💬 Community: https://discord.gg/kthulu\n")
}
