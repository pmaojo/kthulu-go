package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/generator"
	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/resolver"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/parser"
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
		Frontend:    "none",
		Auth:        "jwt",
		Enterprise:  false,
	},
	"monolith": {
		Name:        "Monolith",
		Description: "Full-featured monolithic application",
		Features:    []string{"user", "auth", "organization", "contact", "product"},
		Database:    "postgres",
		Frontend:    "react",
		Auth:        "jwt",
		Enterprise:  false,
	},
	"api-gateway": {
		Name:        "API Gateway",
		Description: "API Gateway with routing and load balancing",
		Features:    []string{"user", "auth", "oauthsso", "realtime"},
		Database:    "postgres",
		Frontend:    "none",
		Auth:        "oauth",
		Enterprise:  true,
	},
	"fintech": {
		Name:        "FinTech",
		Description: "Financial services with compliance and security",
		Features:    []string{"user", "auth", "organization", "contact", "product", "invoice", "payment", "verifactu", "audit"},
		Database:    "postgres",
		Frontend:    "react",
		Auth:        "both",
		Enterprise:  true,
	},
	"ecommerce": {
		Name:        "E-commerce",
		Description: "Complete e-commerce platform",
		Features:    []string{"user", "auth", "organization", "contact", "product", "inventory", "invoice", "payment", "notification", "calendar"},
		Database:    "postgres",
		Frontend:    "react",
		Auth:        "oauth",
		Enterprise:  false,
	},
	"saas": {
		Name:        "SaaS Platform",
		Description: "Multi-tenant SaaS application",
		Features:    []string{"user", "auth", "organization", "contact", "product", "invoice", "payment", "oauthsso", "notification", "audit", "realtime"},
		Database:    "postgres",
		Frontend:    "react",
		Auth:        "both",
		Enterprise:  true,
	},
}

var newCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "🚀 Create a new enterprise-ready Kthulu project with intelligent dependency resolution",
	Long: `Create a production-ready Go application with enterprise features and smart module selection.

The create command uses intelligent dependency resolution to analyze your requirements and automatically
include all necessary modules, detect conflicts, and suggest optimizations.

Templates:
  microservice    - Lightweight microservice (default)
  monolith        - Full-featured monolithic application
  api-gateway     - API Gateway with routing
  fintech         - Financial services with compliance
  ecommerce       - E-commerce platform
  saas            - Multi-tenant SaaS application

Examples:
  kthulu create my-app                          # Create with microservice template
  kthulu create my-shop --template=ecommerce    # Use e-commerce template
  kthulu create my-api --features=user,product  # Custom features
  kthulu create my-fintech --enterprise         # Enable enterprise features
  kthulu create my-frontend --frontend=react    # Include React frontend
  
Advanced Features:
  • Intelligent dependency resolution
  • Conflict detection and resolution
  • Performance optimization suggestions
  • Enterprise security patterns
  • Observability integration
  • Multi-frontend support (React, Templ+HTMX, Fyne)`,
	Args: cobra.ExactArgs(1),
	Run:  runNewProjectIntelligent,
}

var (
	newTemplate      string
	newFeatures      []string
	newDatabase      string
	newFrontend      string
	newAuth          string
	newModulePath    string
	newEnterprise    bool
	newObservability bool
	newOutputPath    string
	newDryRun        bool
	newInteractive   bool
)

func init() {
	newCmd.Flags().StringVarP(&newTemplate, "template", "t", "microservice", "Project template")
	newCmd.Flags().StringSliceVarP(&newFeatures, "features", "f", []string{}, "Comma-separated list of features/modules")
	newCmd.Flags().StringVarP(&newDatabase, "database", "d", "", "Database type (sqlite, postgres, mysql)")
	newCmd.Flags().StringVar(&newFrontend, "frontend", "", "Frontend type (react, templ, fyne, none)")
	newCmd.Flags().StringVar(&newAuth, "auth", "", "Auth type (jwt, oauth, both)")
	newCmd.Flags().StringVar(&newModulePath, "module-path", "", "Go module path (default: project name)")
	newCmd.Flags().BoolVar(&newEnterprise, "enterprise", false, "Enable enterprise features")
	newCmd.Flags().BoolVar(&newObservability, "observability", false, "Enable observability stack")
	newCmd.Flags().StringVarP(&newOutputPath, "output", "o", "", "Output directory (default: current directory)")
	newCmd.Flags().BoolVar(&newDryRun, "dry-run", false, "Show what would be generated without creating files")
	newCmd.Flags().BoolVar(&newInteractive, "interactive", false, "Interactive project configuration")

	rootCmd.AddCommand(newCmd)
}

func runNewProjectIntelligent(cmd *cobra.Command, args []string) {
	projectName := args[0]

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

	// Step 8: Run go mod tidy
	if err := runGoModTidy(structure.RootPath); err != nil {
		fmt.Printf("❌ Error running go mod tidy: %v\n", err)
		// Decide if you want to exit here or just warn the user
	}

	// Step 9: Display success message and next steps
	displaySuccessMessage(projectName, config, structure)
}

func runGoModTidy(projectPath string) error {
	fmt.Println("\n🧹 Running go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func buildProjectConfig(projectName string) (*generator.GeneratorConfig, error) {
	// Start with template defaults
	template, exists := projectTemplates[newTemplate]
	if !exists {
		return nil, fmt.Errorf("unknown template: %s", newTemplate)
	}

	config := &generator.GeneratorConfig{
		ProjectName:   projectName,
		OutputPath:    getOutputPath(projectName),
		Frontend:      template.Frontend,
		Database:      template.Database,
		Auth:          template.Auth,
		Features:      template.Features,
		Enterprise:    template.Enterprise,
		Observability: false,
		CustomValues:  make(map[string]string),
	}

	if newModulePath != "" {
		config.CustomValues["module_path"] = newModulePath
	}

	// Override with command flags
	if len(newFeatures) > 0 {
		config.Features = newFeatures
	}
	if newDatabase != "" {
		config.Database = newDatabase
	}
	if newFrontend != "" {
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
	if newOutputPath != "" {
		return filepath.Join(newOutputPath, projectName)
	}

	pwd, _ := os.Getwd()
	return filepath.Join(pwd, projectName)
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

	fmt.Printf("   go run cmd/migrate/main.go  # Run database migrations\n")
	fmt.Printf("   go run cmd/server/main.go   # Start development server\n")

	if config.Frontend == "react" {
		fmt.Printf("\n💻 Frontend development:\n")
		fmt.Printf("   cd frontend\n")
		fmt.Printf("   npm install\n")
		fmt.Printf("   npm run dev\n")
	}

	fmt.Printf("\n🔧 Additional commands:\n")
	fmt.Printf("   kthulu add module <name>    # Add new modules\n")
	fmt.Printf("   kthulu ai suggest          # Get AI recommendations\n")
	fmt.Printf("   kthulu analyze             # Analyze project structure\n")

	fmt.Printf("\n📚 Documentation: https://docs.kthulu.dev\n")
	fmt.Printf("💬 Community: https://discord.gg/kthulu\n")
}
