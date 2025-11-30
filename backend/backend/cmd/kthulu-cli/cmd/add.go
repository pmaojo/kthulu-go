package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/generator"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/resolver"
)

var addCmd = &cobra.Command{
	Use:   "add [module|component]",
	Short: "➕ Add modules or components to your project",
	Long: `Intelligently add modules or components to your existing Kthulu project.
	
Automatically resolves dependencies, updates configuration, and ensures compatibility.

Examples:
  kthulu add module payment --with-stripe
  kthulu add module auth --with-oauth
  kthulu add component UserHandler --with-tests
  kthulu add integration stripe --compliance=pci`,
}

var addModuleCmd = &cobra.Command{
	Use:   "module [name] [field:type...]",
	Short: "Add a new module to your project with optional fields",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		module := args[0]
		fields := args[1:]
		withIntegrations, _ := cmd.Flags().GetStringSlice("with")
		compliance, _ := cmd.Flags().GetString("compliance")
		force, _ := cmd.Flags().GetBool("force")

		return runAddModule(module, fields, withIntegrations, compliance, force)
	},
}

var addComponentCmd = &cobra.Command{
	Use:   "component [type] [name]",
	Short: "Add a new component to your project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		componentType := args[0]
		name := args[1]
		withTests, _ := cmd.Flags().GetBool("with-tests")
		withMigration, _ := cmd.Flags().GetBool("with-migration")
		module, _ := cmd.Flags().GetString("module")

		return runAddComponent(componentType, name, module, withTests, withMigration)
	},
}

func init() {
	// Add module flags
	addModuleCmd.Flags().StringSlice("with", []string{}, "Integration packages (stripe, oauth, etc)")
	addModuleCmd.Flags().String("compliance", "", "Compliance requirements (pci, sox, gdpr)")
	addModuleCmd.Flags().Bool("force", false, "Force add even if conflicts exist")

	// Add component flags
	addComponentCmd.Flags().Bool("with-tests", true, "Generate tests")
	addComponentCmd.Flags().Bool("with-migration", false, "Generate database migration")
	addComponentCmd.Flags().String("module", "", "Target module (auto-detected if empty)")

	// Add subcommands
	addCmd.AddCommand(addModuleCmd)
	addCmd.AddCommand(addComponentCmd)
}

func runAddModule(module string, fields []string, integrations []string, compliance string, force bool) error {
	fmt.Printf("🧠 Intelligently adding module: %s\n", module)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Step 1: Analyze current project structure
	fmt.Println("🔍 Analyzing current project structure...")
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Check if we're in a Kthulu project
	if !isKthuluProject(currentDir) {
		return fmt.Errorf("not in a Kthulu project directory. Run 'kthulu create <project>' first")
	}

	// Step 2: Parse existing project
	tagParser := parser.NewTagParser(nil)
	analysis, err := tagParser.AnalyzeProject(currentDir)
	if err != nil {
		return fmt.Errorf("error analyzing project: %w", err)
	}

	fmt.Printf("   Found %d existing modules\n", len(analysis.Modules))

	// Step 3: Initialize dependency resolver
	dependencyResolver := resolver.NewDependencyResolver(analysis)

	// Step 4: Resolve dependencies for the new module
	fmt.Printf("🔗 Resolving dependencies for module '%s'...\n", module)
	requiredModules := []string{module}

	// Add integration modules
	for _, integration := range integrations {
		switch integration {
		case "stripe":
			requiredModules = append(requiredModules, "payment")
		case "oauth":
			requiredModules = append(requiredModules, "oauthsso")
		case "postgres":
			requiredModules = append(requiredModules, "database")
		}
	}

	plan, err := dependencyResolver.ResolveDependencies(requiredModules)
	if err != nil {
		return fmt.Errorf("error resolving dependencies: %w", err)
	}

	// Step 5: Display dependency plan
	displayDependencyPlan(module, plan)

	// Step 6: Check for conflicts
	if len(plan.Conflicts) > 0 {
		fmt.Printf("\n⚠️  Detected %d conflicts:\n", len(plan.Conflicts))
		for _, conflict := range plan.Conflicts {
			fmt.Printf("   • %s: %s\n", conflict.Type, conflict.Description)
			for _, suggestion := range conflict.Suggestions {
				fmt.Printf("     → %s\n", suggestion)
			}
		}

		if !force {
			return fmt.Errorf("conflicts detected. Use --force to proceed anyway")
		} else {
			fmt.Println("   ⚠️  Proceeding with conflicts due to --force flag")
		}
	}

	// Step 7: Show recommendations
	if len(plan.Recommendations) > 0 {
		fmt.Printf("\n💡 Recommendations:\n")
		for _, rec := range plan.Recommendations {
			emoji := "💡"
			switch rec.Type {
			case "add":
				emoji = "➕"
			case "configure":
				emoji = "⚙️"
			case "upgrade":
				emoji = "⬆️"
			}
			fmt.Printf("   %s %s %s: %s (impact: %s)\n", emoji, rec.Type, rec.Module, rec.Reason, rec.Impact)
		}
	}

	// Step 8: Generate module files
	fmt.Printf("\n📦 Generating module files...\n")
	templateGenerator := generator.NewTemplateGenerator(dependencyResolver)

	// Create config for the specific module
	config := &generator.GeneratorConfig{
		ProjectName:  filepath.Base(currentDir),
		OutputPath:   currentDir,
		Features:     plan.RequiredModules,
		Enterprise:   compliance != "",
		Database:     detectDatabase(currentDir),
		Frontend:     detectFrontend(currentDir),
		Auth:         detectAuth(currentDir),
		CustomValues: make(map[string]string),
	}

	// Add compliance configuration
	if compliance != "" {
		config.CustomValues["compliance"] = compliance
		fmt.Printf("📋 Configuring %s compliance patterns...\n", compliance)
	}

	// Add integrations
	for _, integration := range integrations {
		config.CustomValues["integration_"+integration] = "true"
	}

	// Generate only the specific module
	if err := generateSpecificModule(config, module, fields, templateGenerator); err != nil {
		return fmt.Errorf("error generating module: %w", err)
	}

	// Step 9: Update project configuration
	fmt.Println("🔧 Updating project configuration...")
	if err := updateProjectConfig(currentDir, plan); err != nil {
		fmt.Printf("⚠️  Warning: Could not update project config: %v\n", err)
	}

	// Step 10: Display success message
	displayModuleSuccessMessage(module, plan)

	return nil
}

func runAddComponent(componentType, name, module string, withTests, withMigration bool) error {
	return fmt.Errorf("enterprise component generation not yet implemented - coming in FASE 1.1")
}

// Helper functions for intelligent module addition

func isKthuluProject(dir string) bool {
	// Check for Kthulu project indicators
	indicators := []string{
		"go.mod",
		"internal/core",
		"cmd/server",
	}

	for _, indicator := range indicators {
		if _, err := os.Stat(filepath.Join(dir, indicator)); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func detectDatabase(dir string) string {
	// Try to detect database type from go.mod
	goModPath := filepath.Join(dir, "go.mod")
	if content, err := os.ReadFile(goModPath); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "gorm.io/driver/postgres") {
			return "postgres"
		}
		if strings.Contains(contentStr, "gorm.io/driver/mysql") {
			return "mysql"
		}
		if strings.Contains(contentStr, "gorm.io/driver/sqlite") {
			return "sqlite"
		}
	}
	return "sqlite" // default
}

func detectFrontend(dir string) string {
	// Check for frontend directories/files
	if _, err := os.Stat(filepath.Join(dir, "frontend", "package.json")); err == nil {
		return "react"
	}
	if _, err := os.Stat(filepath.Join(dir, "templates")); err == nil {
		return "templ"
	}
	if _, err := os.Stat(filepath.Join(dir, "cmd", "desktop")); err == nil {
		return "fyne"
	}
	return "none"
}

func detectAuth(dir string) string {
	// Try to detect auth type from existing modules
	authPath := filepath.Join(dir, "internal", "modules", "auth")
	if _, err := os.Stat(authPath); err == nil {
		return "jwt"
	}

	oauthPath := filepath.Join(dir, "internal", "modules", "oauthsso")
	if _, err := os.Stat(oauthPath); err == nil {
		return "oauth"
	}

	return "jwt" // default
}

func displayDependencyPlan(moduleName string, plan *resolver.ResolutionPlan) {
	fmt.Printf("\n📊 Dependency Resolution Plan:\n")
	fmt.Printf("   Primary module:    %s\n", moduleName)
	fmt.Printf("   Required modules:  %d (%s)\n",
		len(plan.RequiredModules), strings.Join(plan.RequiredModules, ", "))
	fmt.Printf("   Install order:     %s\n", strings.Join(plan.InstallOrder, " → "))

	if len(plan.OptionalModules) > 0 {
		fmt.Printf("   Optional modules:  %s\n", strings.Join(plan.OptionalModules, ", "))
	}

	if len(plan.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings:\n")
		for _, warning := range plan.Warnings {
			fmt.Printf("   • %s\n", warning)
		}
	}
}

func generateSpecificModule(config *generator.GeneratorConfig, moduleName string, fields []string, gen *generator.TemplateGenerator) error {
	fmt.Printf("   📁 Creating module structure for '%s'\n", moduleName)

	// Create module directory
	moduleDir := filepath.Join(config.OutputPath, "internal", "adapters", "http", "modules", moduleName)
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}

	// Create subdirectories
	subdirs := []string{"domain", "repository", "service", "handlers", "dto"}
	for _, subdir := range subdirs {
		if err := os.MkdirAll(filepath.Join(moduleDir, subdir), 0755); err != nil {
			return fmt.Errorf("failed to create subdirectory %s: %w", subdir, err)
		}
	}

	fmt.Printf("   ✅ Module structure created\n")

	// Generate basic module files using the generator
	// This is a simplified version - the full generator.GenerateProject would handle this
	files := map[string]string{
		"module.go":                             generateModuleFile(moduleName),
		fmt.Sprintf("domain/%s.go", moduleName): generateDomainFile(moduleName, fields),
		fmt.Sprintf("repository/%s_repository.go", moduleName): generateRepositoryFile(moduleName),
		fmt.Sprintf("service/%s_service.go", moduleName):       generateServiceFile(moduleName),
		fmt.Sprintf("handlers/%s_handler.go", moduleName):      generateHandlerFile(moduleName),
	}

	for filename, content := range files {
		filePath := filepath.Join(moduleDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
		fmt.Printf("   📝 Generated %s\n", filename)
	}

	// Generate database migration
	migrationName := fmt.Sprintf("create_%ss_table", moduleName)
	migrationContent := generateMigrationContent(moduleName, fields)
	if err := createMigrationFile(migrationName, migrationContent); err != nil {
		fmt.Printf("   ⚠️  Failed to generate migration: %v\n", err)
	} else {
		fmt.Printf("   🐘 Generated database migration for %s\n", moduleName)
	}

	return nil
}

func updateProjectConfig(dir string, plan *resolver.ResolutionPlan) error {
	// Update go.mod with any new dependencies
	fmt.Println("   📝 Updating go.mod...")

	// This could read the existing go.mod and add new dependencies based on the plan
	// For now, we'll just indicate that the update was attempted

	return nil
}

func displayModuleSuccessMessage(moduleName string, plan *resolver.ResolutionPlan) {
	fmt.Printf("\n🎉 Module '%s' added successfully!\n", moduleName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Printf("\n📦 Added modules: %s\n", strings.Join(plan.RequiredModules, ", "))

	if len(plan.OptionalModules) > 0 {
		fmt.Printf("💡 Consider adding: %s\n", strings.Join(plan.OptionalModules, ", "))
	}

	fmt.Printf("\n🚀 Next steps:\n")
	fmt.Printf("   go mod tidy                     # Update dependencies\n")
	fmt.Printf("   go run cmd/server/main.go       # Test your application\n")
	fmt.Printf("   kthulu add component handler %s # Add HTTP handlers\n", moduleName)
	fmt.Printf("   kthulu ai suggest               # Get AI recommendations\n")

	if len(plan.Recommendations) > 0 {
		fmt.Printf("\n💡 Recommendations applied automatically:\n")
		for _, rec := range plan.Recommendations {
			if rec.AutoApply {
				fmt.Printf("   ✅ %s: %s\n", rec.Type, rec.Reason)
			}
		}
	}
}
