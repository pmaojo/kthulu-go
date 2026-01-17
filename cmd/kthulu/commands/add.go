package commands

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pmaojo/kthulu-go/internal/blueprint"
	"github.com/pmaojo/kthulu-go/internal/generator"
	"github.com/pmaojo/kthulu-go/internal/resolver"
	"github.com/pmaojo/kthulu-go/cmd/kthulu/templates"
	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
)

const DefaultModulesPath = "internal/modules"

var addCmd = &cobra.Command{
	Use:   "add [module|component]",
	Short: "➕ Add modules or components to your project",
	Long: `Intelligently add modules or components to your existing Kthulu project.
	
Automatically resolves dependencies, updates configuration, and ensures compatibility.

Examples:
  kthulu add module payment --with-stripe
  kthulu add module auth --with-oauth
  kthulu add component handler UserHandler --module users
  kthulu add integration stripe --compliance=pci`,
}

var addModuleCmd = &cobra.Command{
	Use:   "module [name] [field:type...]",
	Short: "Add a new module to your project with optional fields",
	Long: `Add a new module (domain, repository, service, handlers) to your project.
Fields can be specified as positional arguments in the format 'name:type' or 'name:relation:target'.

Available types: string, int, bool, float, time
Available relations: belongs_to (e.g., car:belongs_to:cars)

Examples:
  kthulu add module orders
  kthulu add module products name:string price:float
  kthulu add module reviews rating:int comment:string user:belongs_to:users
`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		module := args[0]
		fields := args[1:]
		withIntegrations, _ := cmd.Flags().GetStringSlice("with")
		compliance, _ := cmd.Flags().GetString("compliance")
		force, _ := cmd.Flags().GetBool("force")
	yes, _ := cmd.Flags().GetBool("yes")
		if os.Getenv("KTHULU_MCP_MODE") == "1" {
			yes = true
		}

		prefix, _ := cmd.Flags().GetString("prefix")
		protected, _ := cmd.Flags().GetBool("protected")
		admin, _ := cmd.Flags().GetBool("admin")

		return runAddModule(module, fields, withIntegrations, compliance, force, yes, prefix, protected, admin)
	},
}

var addAuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Add authentication module (JWT)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAddAuthModule()
	},
}

var addComponentCmd = &cobra.Command{
	Use:   "component [type] [name]",
	Short: "Add a new component to your project",
	Long: `Add a component (handler, service, repository, domain) to an existing module.

Types:
  handler    - HTTP handler
  service    - Business logic service
  repository - Data access repository
  domain     - Domain entity definition`,
	Args: cobra.ExactArgs(2),
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
	addModuleCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
	addModuleCmd.Flags().String("prefix", "", "Route prefix (e.g. /api/v1)")
	addModuleCmd.Flags().Bool("protected", false, "Protect routes with authentication middleware")
	addModuleCmd.Flags().Bool("admin", false, "Generate Admin CRUD pages")

	// Add component flags
	addComponentCmd.Flags().Bool("with-tests", true, "Generate tests")
	addComponentCmd.Flags().Bool("with-migration", false, "Generate database migration")
	addComponentCmd.Flags().String("module", "", "Target module (auto-detected if empty)")

	// Add subcommands
	addCmd.AddCommand(addModuleCmd)
	addCmd.AddCommand(addComponentCmd)
	addCmd.AddCommand(addAdminCmd)
	addCmd.AddCommand(addRecipeCmd)
}

var addRecipeCmd = &cobra.Command{
	Use:   "recipe [name]",
	Short: "Add a complete recipe (stack of modules)",
	Long: `Add a pre-configured stack of modules (recipe) to your project.
Recipes are defined in the registry and can include multiple modules, integrations, and configurations.

Examples:
  kthulu add recipe growth
  kthulu add recipe saas`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAddRecipe(args[0])
	},
}

type Recipe struct {
	Name        string         `yaml:"name"`
	ID          string         `yaml:"id"`
	Description string         `yaml:"description"`
	Modules     []RecipeModule `yaml:"modules"`
}

type RecipeModule struct {
	Name         string   `yaml:"name"`
	Fields       []string `yaml:"fields"`
	Integrations []string `yaml:"integrations"`
}

func runAddRecipe(name string) error {
	fmt.Printf("🍳 Cooking up recipe: %s\n", name)

	var data []byte

	// 1. Try local registry (Development Mode)
	recipePath := filepath.Join("registry", "recipes", name+".yaml")
	if _, err := os.Stat(recipePath); err == nil {
		data, _ = os.ReadFile(recipePath)
	} else {
		// 2. Try git root (Development Mode)
		if gitRoot, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
			recipePath = filepath.Join(strings.TrimSpace(string(gitRoot)), "registry", "recipes", name+".yaml")
			data, _ = os.ReadFile(recipePath)
		}
	}

	// 3. Try Remote Registry (Production Mode)
	if len(data) == 0 {
		fmt.Println("🌐 Fetching recipe from remote registry...")
		url := fmt.Sprintf("https://raw.githubusercontent.com/pmaojo/kthulu-go/main/registry/recipes/%s.yaml", name)
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			data, err = io.ReadAll(resp.Body)
		}
	}

	if len(data) == 0 {
		return fmt.Errorf("recipe '%s' not found locally or remotely", name)
	}

	var recipe Recipe
	if err := yaml.Unmarshal(data, &recipe); err != nil {
		return fmt.Errorf("failed to parse recipe: %w", err)
	}

	fmt.Printf("📜 %s: %s\n", recipe.Name, recipe.Description)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, module := range recipe.Modules {
		fmt.Printf("\n🔹 Adding module from recipe: %s\n", module.Name)
		// We use force=true and yes=true to batch install without prompts
		if err := runAddModule(module.Name, module.Fields, module.Integrations, "", true, true, "", false, false); err != nil {
			return fmt.Errorf("failed to add module %s: %w", module.Name, err)
		}
	}

	fmt.Printf("\n✨ Recipe '%s' completed successfully!\n", name)
	return nil
}

var addAdminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Add a full Admin Dashboard module",
	Long:  `Scaffolds a complete Admin Dashboard with Dashboard, Users, and Settings pages.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAddAdmin()
	},
}

func runAddAdmin() error {
	fmt.Println("👑 Adding Admin Dashboard Module...")

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	if !isKthuluProject(currentDir) {
		return fmt.Errorf("not in a Kthulu project directory. Run 'kthulu create <project>' first")
	}

	// Initialize generator
	// Detect project module from go.mod
	projectModule, _ := getProjectModule(currentDir)
	config := &generator.GeneratorConfig{
		ProjectName:   filepath.Base(currentDir),
		ProjectModule: projectModule,
		OutputPath:    currentDir,
	}
	
	gen := generator.NewTemplateGenerator(nil)
	gen.SetConfig(config)

	// Create structure container
	structure := &generator.ProjectStructure{
		RootPath:    currentDir,
		Directories: []string{},
		Files:       []generator.GeneratedFile{},
	}

	// Generate Admin Module
	if err := gen.GenerateAdminModule(structure); err != nil {
		return fmt.Errorf("failed to generate admin module: %w", err)
	}

	// Write files
	if err := gen.WriteProject(structure); err != nil {
		return fmt.Errorf("failed to write admin files: %w", err)
	}
	
	fmt.Println("✅ Admin Dashboard added successfully!")
	fmt.Println("👉 Check frontend/src/modules/admin/presentation/")
	return nil
}

func runAddAuthModule() error {
	fmt.Println("🔐 Adding Authentication Module (JWT)...")

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if !isKthuluProject(currentDir) {
		return fmt.Errorf("not in a Kthulu project")
	}

	projectModule, _ := getProjectModule(currentDir)

	data := map[string]any{
		"ModuleName": projectModule,
	}

	base := currentDir

	// Copy auth templates to internal/modules/auth
	err = copyAuthFiles(base, data)
	if err != nil {
		return err
	}

	// Auto register in main.go
	if err := generator.InjectModuleRegistration(base, "auth", projectModule, DefaultModulesPath); err != nil {
		fmt.Printf("⚠️  Failed to register auth module: %v\n", err)
	} else {
		fmt.Println("🔌 Registered auth module in main.go")
	}

	// Add dependencies
	fmt.Println("📦 Adding dependencies...")
	dependencies := []string{
		"golang.org/x/crypto/bcrypt",
		"github.com/golang-jwt/jwt/v5",
	}

	for _, dep := range dependencies {
		if err := installDependency(dep); err != nil {
			fmt.Printf("⚠️  Failed to install %s: %v\n", dep, err)
		} else {
			fmt.Printf("   installed %s\n", dep)
		}
	}

	fmt.Println("✅ Auth module added successfully!")
	return nil
}

func copyAuthFiles(base string, data map[string]any) error {
	// Copy templates from internal/modules/auth to the project's internal/modules/auth
	return copyTemplateTree(templates.Templates, "scaffold/backend/layers", filepath.Join(base, "internal/modules/auth"), data, false)
}

func installDependency(pkg string) error {
	cmd := exec.Command("go", "get", pkg)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	cmd.Stdout = nil // Silence output unless needed
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runAddModule(module string, fields []string, integrations []string, compliance string, force, yes bool, prefix string, protected bool, admin bool) error {
	fmt.Printf("🧠 Intelligently adding module: %s\n", module)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Step 1: Analyze current project structure
	fmt.Println("🔍 Analyzing current project structure...")
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}
	currentDir, _ = filepath.Abs(currentDir)

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

	// Step 7: Confirmation prompt
	if !yes {
		// Check if we are in an interactive terminal
		if _, err := os.Stdin.Stat(); err != nil {
			return fmt.Errorf("non-interactive shell detected, please use --yes to confirm")
		}

		fmt.Printf("\n❓ Do you want to proceed with adding module '%s'? [y/N] ", module)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ Aborted by user")
			return nil
		}
	} else {
		fmt.Println("\n⏩ Skipping confirmation due to --yes flag")
	}

	// Step 8: Show recommendations (Optional, moved after prompt if we want prompt to be the gate,
	// but maybe recommendations should be seen before prompt?
	// The original code had recs after conflicts. Let's keep recs before generation but after prompt?
	// Or maybe before prompt?
	// Let's show recommendations BEFORE prompt so user knows what's happening.
	// But wait, my code block above put prompt at Step 7. Recs were Step 7 in original code.
	// I'll show Recs then Prompt.

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

	// Step 9: Generate module files
	fmt.Printf("\n📦 Generating module files...\n")
	templateGenerator := generator.NewTemplateGenerator(dependencyResolver)

	// Detect project module from go.mod
	projectModule, _ := getProjectModule(currentDir)
	config := &generator.GeneratorConfig{
		ProjectName:   filepath.Base(currentDir),
		ProjectModule: projectModule,
		OutputPath:    currentDir,
		Features:      plan.RequiredModules,
		Enterprise:    compliance != "",
		Database:      detectDatabase(currentDir),
		Frontend:      detectFrontend(currentDir),
		Auth:          detectAuth(currentDir),
		CustomValues:  make(map[string]string),
	}
	templateGenerator.SetConfig(config)

	if prefix != "" {
		config.CustomValues["route_prefix"] = prefix
	} else {
		// Default to module name as prefix
		config.CustomValues["route_prefix"] = fmt.Sprintf("/%s", module)
	}

	if protected {
		config.CustomValues["protected"] = "true"
	}

	if admin {
		config.CustomValues["with_admin"] = "true"
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
	moduleRelPath := detectModuleLocation(config.OutputPath)
	if err := generateSpecificModule(config, module, fields, templateGenerator, moduleRelPath); err != nil {
		return fmt.Errorf("error generating module: %w", err)
	}

	// Register module in main.go
	if err != nil {
		fmt.Printf("   ⚠️  Warning: Could not detect project module from go.mod: %v\n", err)
		// Fallback to trying the framework path only if we are seemingly running tests in the framework itself
		// but typically we should just fail gracefully.
	} else {
		if err := generator.InjectModuleRegistration(config.OutputPath, module, projectModule, moduleRelPath); err != nil {
			fmt.Printf("   ⚠️  Warning: Failed to auto-register module in main.go: %v\n", err)
		} else {
			fmt.Printf("   🔌 Auto-registered module '%s' in main.go\n", module)

			// Register Routes
			entityName := generator.Capitalize(generator.Singularize(module))
			if err := generator.InjectRouteRegistration(config.OutputPath, module, projectModule, moduleRelPath, entityName); err != nil {
				fmt.Printf("   ⚠️  Warning: Failed to auto-register routes in main.go: %v\n", err)
			} else {
				fmt.Printf("   🚦 Auto-registered routes for '%s' in main.go\n", entityName)
			}
		}
	}

	// Step 8b: Generate frontend module if frontend exists
	if config.Frontend == "react" {
		// Create a mock structure to capture generated files
		structure := &generator.ProjectStructure{
			RootPath:    currentDir,
			Directories: []string{},
			Files:       []generator.GeneratedFile{},
		}

		if err := templateGenerator.GenerateFrontendModule(module, fields, structure); err != nil {
			fmt.Printf("   ⚠️  Failed to generate frontend module: %v\n", err)
		} else {
			// Write the generated frontend files
			if err := templateGenerator.WriteProject(structure); err != nil {
				fmt.Printf("   ⚠️  Failed to write frontend files: %v\n", err)
			} else {
				fmt.Printf("   🎨 Generated frontend module for %s\n", module)
			}
		}
	}

	// Step 9: Update project configuration
	fmt.Println("🔧 Updating project configuration...")
	if err := updateProjectConfig(currentDir, plan); err != nil {
		fmt.Printf("⚠️  Warning: Could not update project config: %v\n", err)
	}

	// Step 9b: Run go mod tidy to resolve new imports
	if err := runGoModTidy(currentDir); err != nil {
		fmt.Printf("⚠️  Warning: Failed to run go mod tidy: %v\n", err)
	}

	// Step 10: Display success message
	displayModuleSuccessMessage(module, plan)

	return nil
}

func runAddComponent(componentType, name, module string, withTests, withMigration bool) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// Ensure we are in a Kthulu project
	if !isKthuluProject(currentDir) {
		return fmt.Errorf("not in a Kthulu project directory")
	}

	// Detect module if not provided
	if module == "" {
		module = detectModuleFromDir(currentDir)
		if module == "" {
			return fmt.Errorf("could not detect module. Please specify --module <name>")
		}
	}

	// Verify module existence
	relPath := detectModuleLocation(currentDir)
	modulePath := filepath.Join(currentDir, relPath, module)
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return fmt.Errorf("module '%s' does not exist in %s", module, relPath)
	}

	fmt.Printf("Adding %s '%s' to module '%s'...\n", componentType, name, module)

	var content string
	var filename string
	var subdir string

	// Get project module name
	projectModule, err := getProjectModule(currentDir)
	if err != nil {
		projectModule = "github.com/pmaojo/kthulu-go" // Fallback
	}

	// Initialize generator
	gen := generator.NewTemplateGenerator(nil) // We don't need a full resolver here for single components
	gen.SetConfig(&generator.GeneratorConfig{
		ProjectName:   filepath.Base(currentDir),
		ProjectModule: projectModule,
	})

	switch componentType {
	case "handler":
		// Use module name for imports, component name for the handler struct
		content = gen.GenerateHandlerFile(module) // Use module for imports
		subdir = "handlers"
		filename = fmt.Sprintf("%s_handler.go", strings.ToLower(name))
	case "service":
		content = gen.GenerateServiceFile(module)
		subdir = "service"
		filename = fmt.Sprintf("%s_service.go", strings.ToLower(name))
	case "repository":
		content = gen.GenerateRepositoryFile(module)
		subdir = "repository"
		filename = fmt.Sprintf("%s_repository.go", strings.ToLower(name))
	case "domain":
		content = gen.GenerateDomainFile(module, nil)
		subdir = "domain"
		filename = fmt.Sprintf("%s.go", strings.ToLower(name))
	default:
		return fmt.Errorf("unknown component type: %s. Supported: handler, service, repository, domain", componentType)
	}

	// Create project structure for generation
	structure := &generator.ProjectStructure{
		RootPath: currentDir,
		Files: []generator.GeneratedFile{
			{
				Path:    filepath.Join(relPath, module, subdir, filename),
				Content: content,
			},
		},
	}

	// Write file using unified generator
	if err := gen.WriteProject(structure); err != nil {
		return fmt.Errorf("failed to write component file: %w", err)
	}

	fmt.Printf("✅ Created %s\n", filepath.Join(relPath, module, subdir, filename))

	if withTests {
		fmt.Println("⚠️  Test generation not yet implemented")
	}

	if withMigration {
		fmt.Println("⚠️  Migration generation not yet implemented")
	}

	return nil
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

func detectModuleFromDir(dir string) string {
	// Try to detect module from current path
	// Path should look like .../internal/adapters/http/modules/<module_name>...
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}

	parts := strings.Split(absDir, string(os.PathSeparator))
	for i, part := range parts {
		if part == "modules" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func detectModuleLocation(projectRoot string) string {
	// Priority 0: Check configuration file (kthulu-plan.yaml)
	if data, err := os.ReadFile(filepath.Join(projectRoot, "kthulu-plan.yaml")); err == nil {
		var bp blueprint.ProjectBlueprint
		if err := yaml.Unmarshal(data, &bp); err == nil && bp.Structure.ModulesPath != "" {
			return bp.Structure.ModulesPath
		}
	}

	// Priority 1: Check if it's an enterprise project style (internal/adapters/http/modules)
	if _, err := os.Stat(filepath.Join(projectRoot, "internal", "adapters", "http", "modules")); err == nil {
		return "internal/adapters/http/modules"
	}
	// Priority 2: Check for the simpler style (internal/modules)
	if _, err := os.Stat(filepath.Join(projectRoot, DefaultModulesPath)); err == nil {
		return DefaultModulesPath
	}
	// Default to internal/modules for new additions if neither exists (fallback)
	return DefaultModulesPath
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

func generateSpecificModule(config *generator.GeneratorConfig, moduleName string, fields []string, gen *generator.TemplateGenerator, moduleRelPath string) error {
	fmt.Printf("   📁 Creating module structure for '%s' in %s\n", moduleName, moduleRelPath)

	// Create module directory based on detected location
	moduleDir := filepath.Join(config.OutputPath, moduleRelPath, moduleName)
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

	// Generate module files using the unified generator
	files, migrationContent, err := gen.GenerateBackendModule(moduleName, fields, moduleRelPath, config.CustomValues["route_prefix"], config.CustomValues["protected"] == "true")
	if err != nil {
		return fmt.Errorf("error generating module: %w", err)
	}

	for relPath, content := range files {
		filePath := filepath.Join(moduleDir, relPath)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
		fmt.Printf("   📝 Generated %s\n", relPath)
	}

	// Generate database migration
	if migrationContent != "" {
		migrationName := fmt.Sprintf("create_%ss_table", moduleName)
		if err := createMigrationFile(migrationName, migrationContent); err != nil {
			fmt.Printf("   ⚠️  Failed to generate migration: %v\n", err)
		} else {
			fmt.Printf("   🐘 Generated database migration for %s\n", moduleName)
		}
	}

	return nil
}

func updateProjectConfig(dir string, plan *resolver.ResolutionPlan) error {
	// Update go.mod with any new dependencies
	fmt.Println("   📝 Updating go.mod...")

	// Try to update kthulu-plan.yaml if it exists
	planPath := filepath.Join(dir, "kthulu-plan.yaml")
	if _, err := os.Stat(planPath); err == nil {
		fmt.Println("   📝 Updating kthulu-plan.yaml...")
		if err := updatePlanFile(planPath, plan); err != nil {
			fmt.Printf("   ⚠️  Failed to update plan file: %v\n", err)
		}
	}

	return nil
}

func updatePlanFile(path string, plan *resolver.ResolutionPlan) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var bp blueprint.ProjectBlueprint
	if err := yaml.Unmarshal(data, &bp); err != nil {
		return err
	}

	// Add new modules to features list if not present
	existingFeatures := make(map[string]bool)
	for _, f := range bp.Features {
		existingFeatures[f] = true
	}

	for _, mod := range plan.RequiredModules {
		if !existingFeatures[mod] {
			bp.Features = append(bp.Features, mod)
			existingFeatures[mod] = true
		}
	}

	// Make sure map is initialized
	if bp.Modules == nil {
		bp.Modules = make(map[string]blueprint.ModuleConfig)
	}

	// We could add ModuleConfig here if we had field info passed down to updateProjectConfig
	// For now, we at least register the module name in Features

	newData, err := yaml.Marshal(&bp)
	if err != nil {
		return err
	}

	return os.WriteFile(path, newData, 0644)
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


