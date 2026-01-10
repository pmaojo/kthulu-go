package generator

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"text/template"

	"github.com/jinzhu/inflection"
	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/resolver"
	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/templates"
)

// TemplateGenerator generates code templates based on dependency analysis
type TemplateGenerator struct {
	resolver  *resolver.DependencyResolver
	templates map[string]*template.Template
	config    *GeneratorConfig
}

// GeneratorConfig configures the template generation
type GeneratorConfig struct {
	ProjectName   string            `json:"project_name"`
	OutputPath    string            `json:"output_path"`
	Frontend      string            `json:"frontend"`      // react, templ, fyne, none
	ProjectModule string            `json:"project_module"`
	Database      string            `json:"database"`      // sqlite, postgres, mysql
	Auth          string            `json:"auth"`          // jwt, oauth, both
	Features      []string          `json:"features"`      // modules to include
	Enterprise    bool              `json:"enterprise"`    // enterprise features
	Observability bool              `json:"observability"` // monitoring
	CustomValues    map[string]string   `json:"custom_values"`    // custom template values
	ModuleFields    map[string][]string `json:"module_fields"`    // fields for each module
	FrontendModules []string            `json:"frontend_modules"` // modules that get frontend (from schema 'modules:')
}

// modulePath returns the module import path for the generated project.
func (g *TemplateGenerator) modulePath() string {
	if g.config == nil {
		return ""
	}

	if g.config.ProjectModule != "" {
		return g.config.ProjectModule
	}

	if g.config.CustomValues != nil {
		if modulePath := strings.TrimSpace(g.config.CustomValues["module_path"]); modulePath != "" {
			return modulePath
		}
	}

	return strings.TrimSpace(g.config.ProjectName)
}

// moduleImportPath builds an import path anchored at the module path.
func (g *TemplateGenerator) moduleImportPath(parts ...string) string {
	base := strings.Trim(g.modulePath(), "/")
	segments := make([]string, 0, len(parts)+1)
	if base != "" {
		segments = append(segments, base)
	}

	for _, part := range parts {
		if trimmed := strings.Trim(part, "/"); trimmed != "" {
			segments = append(segments, trimmed)
		}
	}

	if len(segments) == 0 {
		return ""
	}

	return path.Join(segments...)
}

// ProjectStructure represents the generated project structure
type ProjectStructure struct {
	RootPath      string                 `json:"root_path"`
	Directories   []string               `json:"directories"`
	Files         []GeneratedFile        `json:"files"`
	Dependencies  []string               `json:"dependencies"`
	Scripts       map[string]string      `json:"scripts"`
	Configuration map[string]interface{} `json:"configuration"`
}

// GeneratedFile represents a generated file
type GeneratedFile struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	Template   string `json:"template"`
	Executable bool   `json:"executable"`
	Overwrite  bool   `json:"overwrite"`
}

// SetConfig updates the generator configuration
func (g *TemplateGenerator) SetConfig(config *GeneratorConfig) {
	g.config = config
}

// NewTemplateGenerator creates a new template generator
func NewTemplateGenerator(resolver *resolver.DependencyResolver) *TemplateGenerator {
	return &TemplateGenerator{
		resolver:  resolver,
		templates: make(map[string]*template.Template),
		config:    &GeneratorConfig{},
	}
}

func (g *TemplateGenerator) executeTemplate(name, path string, data interface{}) (string, error) {
	tmpl, ok := g.templates[name]
	if !ok {
		content, err := templates.Templates.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read template %s: %w", path, err)
		}

		tmpl, err = template.New(name).Funcs(template.FuncMap{
			"Capitalize":  Capitalize,
			"capitalize":  Capitalize,
			"Pluralize":   Pluralize,
			"pluralize":   Pluralize,
			"ToSnakeCase": ToSnakeCase,
			"ToKebabCase": ToKebabCase,
			"lower":       strings.ToLower,
		}).Parse(string(content))
		if err != nil {
			return "", fmt.Errorf("failed to parse template %s: %w", path, err)
		}
		g.templates[name] = tmpl
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}
	return buf.String(), nil
}

// GenerateProject generates a complete project based on dependency analysis
func (g *TemplateGenerator) GenerateProject(config *GeneratorConfig) (*ProjectStructure, error) {
	fmt.Printf("🏗️  Generating project '%s' with features: %s\n",
		config.ProjectName, strings.Join(config.Features, ", "))

	g.config = config

	// Step 1: Resolve dependencies
	plan, err := g.resolver.ResolveDependencies(config.Features)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	if len(plan.Conflicts) > 0 {
		return nil, fmt.Errorf("dependency conflicts detected: %v", plan.Conflicts)
	}

	// Step 2: Initialize project structure
	structure := &ProjectStructure{
		RootPath:      config.OutputPath,
		Directories:   []string{},
		Files:         []GeneratedFile{},
		Dependencies:  plan.RequiredModules,
		Scripts:       make(map[string]string),
		Configuration: make(map[string]interface{}),
	}

	// Step 3: Generate base structure
	if err := g.generateBaseStructure(structure); err != nil {
		return nil, fmt.Errorf("failed to generate base structure: %w", err)
	}

	// Step 4: Generate module files
	for _, module := range plan.InstallOrder {
		if err := g.generateModuleFiles(module, structure); err != nil {
			return nil, fmt.Errorf("failed to generate module '%s': %w", module, err)
		}
	}

	// Step 5: Generate frontend if requested
	if config.Frontend == "react" {
		if err := g.generateFrontend(structure); err != nil {
			return nil, fmt.Errorf("failed to generate frontend: %w", err)
		}
	}

	// Step 6: Generate configuration files
	if err := g.generateConfiguration(structure); err != nil {
		return nil, fmt.Errorf("failed to generate configuration: %w", err)
	}

	// Step 7: Generate build scripts
	if err := g.generateBuildScripts(structure); err != nil {
		return nil, fmt.Errorf("failed to generate build scripts: %w", err)
	}

	fmt.Printf("✅ Project generated successfully: %d files, %d directories\n",
		len(structure.Files), len(structure.Directories))

	return structure, nil
}

// generateBaseStructure generates the base project structure
func (g *TemplateGenerator) generateBaseStructure(structure *ProjectStructure) error {
	baseDirs := []string{
		"cmd/server",
		"cmd/migrate",
		"internal/core",
		g.getModuleRelPath(),
		"configs",
		"migrations",
		"scripts",
	}

	if g.config.Enterprise {
		baseDirs = append(baseDirs,
			"internal/audit",
			"internal/security",
			"internal/compliance",
			"internal/monitoring",
		)
	}

	structure.Directories = append(structure.Directories, baseDirs...)
	structure.Directories = append(structure.Directories,
		"internal/infrastructure/middleware",
		"internal/infrastructure/observability",
	)

	// Generate main.go
	mainFile := GeneratedFile{
		Path:     "cmd/server/main.go",
		Template: "main.go.tmpl",
		Content:  g.generateMainFile(),
	}
	structure.Files = append(structure.Files, mainFile)

	// Generate main_test.go
	mainTestFile := GeneratedFile{
		Path:    "cmd/server/main_test.go",
		Content: g.generateMainTestFile(),
	}
	structure.Files = append(structure.Files, mainTestFile)

	// Generate go.mod
	goModFile := GeneratedFile{
		Path:     "go.mod",
		Template: "go.mod.tmpl",
		Content:  g.generateGoMod(),
	}
	structure.Files = append(structure.Files, goModFile)

	// Generate README.md
	readmeFile := GeneratedFile{
		Path:     "README.md",
		Template: "README.md.tmpl",
		Content:  g.generateReadme(),
	}
	structure.Files = append(structure.Files, readmeFile)

	coreProviders := GeneratedFile{
		Path:    "internal/core/providers.go",
		Content: g.generateCoreProviders(),
	}
	structure.Files = append(structure.Files, coreProviders)

	coreProvidersTest := GeneratedFile{
		Path:    "internal/core/providers_test.go",
		Content: g.generateCoreProvidersTest(),
	}
	structure.Files = append(structure.Files, coreProvidersTest)

	// Generate Infrastructure Files
	infraFiles := map[string]string{
		"internal/infrastructure/middleware/recovery.go":   "backend/internal/infrastructure/middleware/recovery.go.tmpl",
		"internal/infrastructure/middleware/middleware.go": "backend/internal/infrastructure/middleware/middleware.go.tmpl",
		"internal/infrastructure/observability/logger.go":  "backend/internal/infrastructure/observability/logger.go.tmpl",
		"internal/infrastructure/observability/metrics.go": "backend/internal/infrastructure/observability/metrics.go.tmpl",
	}

	for path, tmpl := range infraFiles {
		content, err := g.executeTemplate(path, tmpl, map[string]interface{}{
			"ModuleName": g.modulePath(),
		})
		if err != nil {
			return fmt.Errorf("failed to generate %s: %w", path, err)
		}
		structure.Files = append(structure.Files, GeneratedFile{
			Path:    path,
			Content: content,
		})
	}

	// Generate cmd/migrate/main.go
	migrateMain := GeneratedFile{
		Path:    "cmd/migrate/main.go",
		Content: g.generateMigrateMainFile(),
	}
	structure.Files = append(structure.Files, migrateMain)

	return nil
}

// generateModuleFiles generates files for a specific module
func (g *TemplateGenerator) generateModuleFiles(moduleName string, structure *ProjectStructure) error {
	fmt.Printf("  📦 Generating module: %s\n", moduleName)

	// Generate module directory structure
	relPath := g.getModuleRelPath()
	moduleBase := fmt.Sprintf("%s/%s", relPath, moduleName)
	moduleDirs := []string{
		moduleBase,
		fmt.Sprintf("%s/domain", moduleBase),
		fmt.Sprintf("%s/repository", moduleBase),
		fmt.Sprintf("%s/service", moduleBase),
		fmt.Sprintf("%s/handlers", moduleBase),
	}
	structure.Directories = append(structure.Directories, moduleDirs...)

	// Generate module files using GenerateBackendModule to ensure consistency
	fields := g.config.ModuleFields[moduleName]
	files, migrationContent, err := g.GenerateBackendModule(moduleName, fields, relPath, "", false)
	if err != nil {
		return err
	}

	for relPath, content := range files {
		structure.Files = append(structure.Files, GeneratedFile{
			Path:    filepath.Join(moduleBase, relPath),
			Content: content,
		})
	}

	// Generate migration
	if migrationContent != "" {
		timestamp := time.Now().Format("20060102150405")
		structure.Files = append(structure.Files, GeneratedFile{
			Path:    filepath.Join("migrations", fmt.Sprintf("%s_create_%s_table.sql", timestamp, moduleName)),
			Content: migrationContent,
		})
	}

	return nil
}

// generateMainFile generates the main.go file
// generateMainFile generates the main.go file
func (g *TemplateGenerator) generateMainFile() string {
	coreImport := g.moduleImportPath("internal/core")
	data := map[string]interface{}{
		"ProjectName":     g.config.ProjectName,
		"Features":        strings.Join(g.config.Features, ","),
		"CoreImport":      coreImport,
		"ModuleImports":   g.generateModuleImports(),
		"ModuleProviders": g.generateModuleProviders(),
		"InvokeParams":    g.generateInvokeParams(),
		"ModuleRoutes":    g.generateModuleRoutes(),
	}

	content, err := g.executeTemplate("main", "scaffold/backend/project/main.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating main.go: %v\n", err)
		return ""
	}
	return content
}

// generateGoMod generates the go.mod file
// generateGoMod generates the go.mod file
func (g *TemplateGenerator) generateGoMod() string {
	modulePath := g.modulePath()
	depSet := make(map[string]struct{})
	var deps []string
	addDep := func(dep string) {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			return
		}
		if _, exists := depSet[dep]; exists {
			return
		}
		depSet[dep] = struct{}{}
		deps = append(deps, "\t"+dep)
	}

	addDep("go.uber.org/fx v1.20.0")
	addDep("github.com/gorilla/mux v1.8.0")
	addDep("gorm.io/gorm v1.25.5")
	if g.config.Database == "sqlite" || g.config.Database == "" {
		addDep("github.com/glebarez/sqlite v1.11.0")
	} else {
		addDep(fmt.Sprintf("gorm.io/driver/%s v1.5.4", g.config.Database))
	}
	addDep("github.com/golang-jwt/jwt/v5 v5.2.0")
	addDep("github.com/pressly/goose/v3 v3.24.3")

	if extra := strings.Split(strings.TrimSpace(g.generateDependencies()), "\n"); len(extra) > 0 {
		for _, dep := range extra {
			addDep(strings.TrimSpace(dep))
		}
	}

	data := map[string]interface{}{
		"ModulePath":   modulePath,
		"Dependencies": strings.Join(deps, "\n"),
	}

	content, err := g.executeTemplate("gomod", "scaffold/backend/project/go.mod.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating go.mod: %v\n", err)
		return ""
	}
	return content
}

// generateReadme generates the README.md file
// generateReadme generates the README.md file
func (g *TemplateGenerator) generateReadme() string {
	data := map[string]interface{}{
		"ProjectName": g.config.ProjectName,
		"FeatureList": g.generateFeatureList(),
		"Database":    g.config.Database,
		"Auth":        g.config.Auth,
		"Features":    strings.Join(g.config.Features, ", "),
		"ModuleCount": len(g.config.Features),
	}

	content, err := g.executeTemplate("readme", "scaffold/backend/project/README.md.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating README.md: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateCoreProviders() string {
	dbDriver := strings.ToLower(strings.TrimSpace(g.config.Database))
	if dbDriver == "" {
		dbDriver = "sqlite"
	}

	dbName := strings.TrimSpace(g.config.ProjectName)
	if dbName == "" {
		dbName = "app"
	}

	imports := []string{
		"\"fmt\"",
		"\"log\"",
		"\"os\"",
		"\"go.uber.org/fx\"",
		"\"gorm.io/gorm\"",
	}

	var connectionBuilder strings.Builder
	var driverImport string
	switch dbDriver {
	case "postgres":
		imports = append(imports, "\"fmt\"")
		driverImport = "\"gorm.io/driver/postgres\""
		connectionBuilder.WriteString("\t\tdsn := fmt.Sprintf(\"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable\",\n")
		connectionBuilder.WriteString("\t\t\tgetEnv(\"DB_HOST\", \"localhost\"),\n")
		connectionBuilder.WriteString("\t\t\tgetEnv(\"DB_PORT\", \"5432\"),\n")
		connectionBuilder.WriteString("\t\t\tgetEnv(\"DB_USER\", \"postgres\"),\n")
		connectionBuilder.WriteString("\t\t\tgetEnv(\"DB_PASSWORD\", \"postgres\"),\n")
		connectionBuilder.WriteString(fmt.Sprintf("\t\t\tgetEnv(\"DB_NAME\", \"%s\"),\n", dbName))
		connectionBuilder.WriteString("\t\t)\n")
		connectionBuilder.WriteString(fmt.Sprintf("\t\tlog.Printf(\"Connecting to PostgreSQL at %%s:%%s/%%s\", getEnv(\"DB_HOST\", \"localhost\"), getEnv(\"DB_PORT\", \"5432\"), getEnv(\"DB_NAME\", \"%s\"))\n", dbName))
		connectionBuilder.WriteString("\t\treturn gorm.Open(postgres.Open(dsn), &gorm.Config{})\n")
	case "mysql":
		imports = append(imports, "\"fmt\"")
		driverImport = "\"gorm.io/driver/mysql\""
		connectionBuilder.WriteString("\t\tdsn := fmt.Sprintf(\"%s:%s@tcp(%s:%s)/%s?parseTime=true\",\n")
		connectionBuilder.WriteString("\t\t\tgetEnv(\"DB_USER\", \"root\"),\n")
		connectionBuilder.WriteString("\t\t\tgetEnv(\"DB_PASSWORD\", \"password\"),\n")
		connectionBuilder.WriteString("\t\t\tgetEnv(\"DB_HOST\", \"localhost\"),\n")
		connectionBuilder.WriteString("\t\t\tgetEnv(\"DB_PORT\", \"3306\"),\n")
		connectionBuilder.WriteString(fmt.Sprintf("\t\t\tgetEnv(\"DB_NAME\", \"%s\"),\n", dbName))
		connectionBuilder.WriteString("\t\t)\n")
		connectionBuilder.WriteString(fmt.Sprintf("\t\tlog.Printf(\"Connecting to MySQL at %%s:%%s/%%s\", getEnv(\"DB_HOST\", \"localhost\"), getEnv(\"DB_PORT\", \"3306\"), getEnv(\"DB_NAME\", \"%s\"))\n", dbName))
		connectionBuilder.WriteString("\t\treturn gorm.Open(mysql.Open(dsn), &gorm.Config{})\n")
	default:
		driverImport = "\"github.com/glebarez/sqlite\""
		imports = append(imports, "\"path/filepath\"")
		connectionBuilder.WriteString(fmt.Sprintf("\t\tdbPath := getEnv(\"SQLITE_PATH\", \"data/%s.db\")\n", dbName))
		connectionBuilder.WriteString("\t\tif err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {\n")
		connectionBuilder.WriteString("\t\t\treturn nil, err\n")
		connectionBuilder.WriteString("\t\t}\n")
		connectionBuilder.WriteString("\t\tlog.Printf(\"Using SQLite database at %s\", dbPath)\n")
		connectionBuilder.WriteString("\t\treturn gorm.Open(sqlite.Open(dbPath), &gorm.Config{})\n")
	}
	imports = append(imports, driverImport)

	// Generate domain imports for AutoMigrate
	var domainImports []string
	var autoMigrateModels []string
	plan, _ := g.resolver.ResolveDependencies(g.config.Features)
	relPath := g.getModuleRelPath()
	for _, module := range plan.RequiredModules {
		domainImport := g.moduleImportPath(relPath, module, "domain")
		domainImports = append(domainImports, fmt.Sprintf("\t%sDomain \"%s\"", module, domainImport))
		autoMigrateModels = append(autoMigrateModels, fmt.Sprintf("&%sDomain.%s{}", module, Capitalize(module)))
	}

	var importLines []string
	for _, imp := range imports {
		importLines = append(importLines, "\t"+imp)
	}
	importLines = append(importLines, domainImports...)

	autoMigrateCall := ""
	if len(autoMigrateModels) > 0 {
		autoMigrateCall = fmt.Sprintf("\n\t// Auto-migrate all domain models\n\tif err := db.AutoMigrate(%s); err != nil {\n\t\treturn nil, fmt.Errorf(\"auto-migrate failed: %%w\", err)\n\t}\n\treturn db, nil", strings.Join(autoMigrateModels, ", "))
	} else {
		autoMigrateCall = "\n\treturn db, nil"
	}

	return fmt.Sprintf(`package core

import (
%s
)

func CoreRepositoryProviders() fx.Option {
        return fx.Options(
                fx.Provide(NewDatabase),
        )
}

func NewDatabase() (*gorm.DB, error) {
if os.Getenv("KTHULU_TEST_MODE") == "1" {
log.Println("Using in-memory SQLite database for tests")
db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
if err != nil { return nil, err }
%s
}
%s
}

func getEnv(key, fallback string) string {
        if value := os.Getenv(key); value != "" {
                return value
        }
        return fallback
}
`, strings.Join(importLines, "\n"), autoMigrateCall, strings.Replace(connectionBuilder.String(), "return gorm.Open(", "db, err := gorm.Open(", 1)+"\nif err != nil { return nil, err }"+autoMigrateCall)
}

// Helper methods for code generation
func (g *TemplateGenerator) generateModuleImports() string {
	var imports []string
	relPath := g.getModuleRelPath()
	// Use resolved dependencies, not just initial features
	plan, _ := g.resolver.ResolveDependencies(g.config.Features)
	for _, module := range plan.RequiredModules {
		moduleBase := g.moduleImportPath(relPath, module)
		domainImport := g.moduleImportPath(relPath, module, "domain")
		handlersImport := g.moduleImportPath(relPath, module, "handlers")
		imports = append(imports, fmt.Sprintf(` "%s"`, moduleBase))
		imports = append(imports, fmt.Sprintf(` %sDomain "%s"`, module, domainImport))
		imports = append(imports, fmt.Sprintf(` %sHandlers "%s"`, module, handlersImport))
	}
	return strings.Join(imports, "\n")
}

func (g *TemplateGenerator) generateModuleProviders() string {
	var providers []string
	plan, _ := g.resolver.ResolveDependencies(g.config.Features)
	for _, module := range plan.RequiredModules {
		providers = append(providers, fmt.Sprintf("\t\t%s.Providers(),", module))
	}
	return strings.Join(providers, "\n")
}

func (g *TemplateGenerator) generateModuleRoutes() string {
	var routes []string
	plan, _ := g.resolver.ResolveDependencies(g.config.Features)
	for _, module := range plan.RequiredModules {
		routes = append(routes, fmt.Sprintf(`	// %s routes`, module))
		routes = append(routes, fmt.Sprintf(`	%sHandler := %sHandlers.New%sHandler(%sService)`, module, module, Capitalize(module), module))
		routes = append(routes, fmt.Sprintf(`	%sHandler.RegisterRoutes(apiRouter)`, module))
	}
	return strings.Join(routes, "\n")
}

func (g *TemplateGenerator) generateInvokeParams() string {
	var params []string
	plan, _ := g.resolver.ResolveDependencies(g.config.Features)
	for _, module := range plan.RequiredModules {
		params = append(params, fmt.Sprintf(`%sService %sDomain.%sService`, module, module, Capitalize(module)))
	}
	return strings.Join(params, ", ")
}

func (g *TemplateGenerator) generateDependencies() string {
	deps := []string{}

	if g.config.Enterprise {
		deps = append(deps,
			"\tgithub.com/prometheus/client_golang v1.17.0",
			"\tgo.opentelemetry.io/otel v1.21.0",
		)
	}

	if g.config.Frontend == "react" {
		deps = append(deps, "\tgithub.com/gorilla/websocket v1.5.0")
	}

	return strings.Join(deps, "\n")
}

func (g *TemplateGenerator) generateFeatureList() string {
	var features []string
	for _, feature := range g.config.Features {
		if info, err := g.resolver.GetModuleInfo(feature); err == nil {
			features = append(features, fmt.Sprintf("- **%s**: %s", info.Name, info.Description))
		} else {
			features = append(features, fmt.Sprintf("- **%s**: Custom module", feature))
		}
	}
	return strings.Join(features, "\n")
}

// getModuleRelPath returns the relative path for modules based on configuration
func (g *TemplateGenerator) getModuleRelPath() string {
	if g.config.Enterprise {
		return "internal/adapters/http/modules"
	}
	return "internal/modules"
}

// Additional generation methods (simplified for brevity)
func (g *TemplateGenerator) GenerateModuleFile(name string, info *resolver.ModuleInfo) string {
	relPath := g.getModuleRelPath()
	data := map[string]interface{}{
		"Name":          name,
		"Title":         Capitalize(inflection.Singular(name)),
		"ProjectModule": g.modulePath(),
		"Module":        name,
		"ModulePath":    g.moduleImportPath(relPath, name),
		"ModuleRelPath": relPath,
	}

	content, err := g.executeTemplate("module", "scaffold/backend/module.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating module file: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) GenerateDomainFile(name string, fields []BackendField) string {
	data := map[string]interface{}{
		"Name":        name,
		"Title":       Capitalize(inflection.Singular(name)),
		"PluralTitle": Pluralize(Capitalize(name)),
		"Fields":      fields,
	}

	content, err := g.executeTemplate("domain", "scaffold/backend/layers/domain.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating domain file: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) GenerateRepositoryFile(name string) string {
	relPath := g.getModuleRelPath()
	data := map[string]interface{}{
		"Name":         name,
		"Title":        Capitalize(inflection.Singular(name)),
		"DomainImport": g.moduleImportPath(relPath, name, "domain"),
	}

	content, err := g.executeTemplate("repository", "scaffold/backend/layers/repository.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating repository file: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) GenerateServiceFile(name string) string {
	relPath := g.getModuleRelPath()
	data := map[string]interface{}{
		"Name":         name,
		"Title":        Capitalize(inflection.Singular(name)),
		"PluralTitle":  Pluralize(Capitalize(name)),
		"DomainImport": g.moduleImportPath(relPath, name, "domain"),
	}

	content, err := g.executeTemplate("service", "scaffold/backend/layers/service.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating service file: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) GenerateHandlerFile(name string) string {
	relPath := g.getModuleRelPath()
	data := map[string]interface{}{
		"Name":         name,
		"Title":        Capitalize(inflection.Singular(name)),
		"PluralTitle":  Pluralize(Capitalize(name)),
		"RoutePrefix":  ToKebabCase(name),
		"DomainImport": g.moduleImportPath(relPath, name, "domain"),
	}

	content, err := g.executeTemplate("handler", "scaffold/backend/layers/handler.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating handler file: %v\n", err)
		return ""
	}
	return content
}

// GenerateBackendModule generates a complete backend module with fields
func (g *TemplateGenerator) GenerateBackendModule(moduleName string, fields []string, moduleRelPath string, routePrefix string, protected bool) (map[string]string, string, error) {
	fmt.Printf("  📦 Generating backend module: %s\n", moduleName)

	backendFields := ParseBackendFields(fields)
	
	// Collect imports
	imports := make([]string, 0)
	seenImports := make(map[string]bool)
	for _, f := range backendFields {
		if f.Relation != "" && f.RelModule != "" {
			alias := ToSnakeCase(f.RelModule) + "Domain"
			impPath := fmt.Sprintf("%s/%s/%s/domain", g.modulePath(), moduleRelPath, f.RelModule)
			imp := fmt.Sprintf(`%s "%s"`, alias, impPath)

			if !seenImports[imp] {
				imports = append(imports, imp)
				seenImports[imp] = true
			}
		}
	}

	data := map[string]interface{}{
		"Name":          moduleName,
		"Title":         Capitalize(inflection.Singular(moduleName)),
		"PluralTitle":   Pluralize(Capitalize(moduleName)),
		"Fields":        backendFields,
		"Imports":       imports,
		"Database":      g.config.Database,
		"ProjectModule": g.modulePath(),
		"ModuleRelPath": moduleRelPath,
		"RoutePrefix":   routePrefix,
		"Protected":     protected,
		"DomainImport":  g.moduleImportPath(moduleRelPath, moduleName, "domain"),
	}

	files := make(map[string]string)
	
	// Core layers
	layers := map[string]string{
		"module.go":                             "scaffold/backend/module.go.tmpl",
		"domain/" + moduleName + ".go":           "scaffold/backend/layers/domain.go.tmpl",
		"repository/" + moduleName + "_repository.go": "scaffold/backend/layers/repository.go.tmpl",
		"service/" + moduleName + "_service.go":       "scaffold/backend/layers/service.go.tmpl",
		"handlers/" + moduleName + "_handler.go":      "scaffold/backend/layers/handler.go.tmpl",
	}

	for relPath, tmplPath := range layers {
		content, err := g.executeTemplate(relPath, tmplPath, data)
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate %s: %w", relPath, err)
		}
		files[relPath] = content
	}

	// Migration
	migrationContent, err := g.executeTemplate("migration", "scaffold/backend/migration.sql.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to generate migration: %v\n", err)
	}

	return files, migrationContent, nil
}
func (g *TemplateGenerator) generateFrontend(structure *ProjectStructure) error {
	// Generate frontend modules for each feature
	// Note: We only generate the basic structure here.
	// Field-specific generation happens in generateFrontendModule
	// which is called by the 'add module' command or when we have field info.

	// For initial project generation, we don't have field info for the modules
	// in g.config.Features, so we skip detailed generation unless we implement
	// a way to pass schema definitions during project creation.

	// However, we can create the directory structure.

	fmt.Println("  🎨 Setting up frontend structure...")
	
	// Generate base scaffold (Vite, React, etc.)
	if err := g.generateFrontendBase(structure); err != nil {
		return err
	}

	for _, feature := range g.config.Features {
		if feature == "auth" || feature == "users" {
			// Skip special modules for now or handle them differently
			continue
		}

		// For standard modules, we create the folder structure AND the basic content.
		// Since we don't have field info at this stage, we'll generate a basic "Hello World"
		// version of the module with just a name field, so the user has something runnable.
		if err := g.GenerateFrontendModule(feature, []string{"name:string"}, structure); err != nil {
			return fmt.Errorf("failed to generate frontend module for %s: %w", feature, err)
		}
	}

	return nil
}

// GenerateFrontendModule generates a specific frontend module with fields
func (g *TemplateGenerator) GenerateFrontendModule(moduleName string, fields []string, structure *ProjectStructure) error {
	fmt.Printf("  🎨 Generating frontend module: %s\n", moduleName)

	frontendFields := ParseFrontendFields(fields)
	data := FrontendTemplateData{
		Name:       moduleName,
		Title:      Capitalize(moduleName),
		PluralName: Pluralize(moduleName),
		Fields:     frontendFields,
	}

	moduleDir := fmt.Sprintf("frontend/src/modules/%s", moduleName)

	// Create directories
	dirs := []string{
		moduleDir,
		path.Join(moduleDir, "domain"),
		path.Join(moduleDir, "infrastructure"),
		path.Join(moduleDir, "application"),
		path.Join(moduleDir, "presentation"),
		path.Join(moduleDir, "presentation", "components"),
	}
	structure.Directories = append(structure.Directories, dirs...)

	// Generate files
	files := map[string]string{
		"domain/" + data.Title + ".ts":                       "scaffold/frontend/react/domain.ts.tmpl",
		"infrastructure/" + data.Title + "Service.ts":        "scaffold/frontend/react/service.ts.tmpl",
		"application/use" + data.Title + "s.ts":              "scaffold/frontend/react/hook.ts.tmpl",
		"presentation/components/" + data.Title + "List.tsx": "scaffold/frontend/react/list.tsx.tmpl",
		"presentation/components/" + data.Title + "Form.tsx": "scaffold/frontend/react/form.tsx.tmpl",
		"presentation/" + data.Title + "Page.tsx":            "scaffold/frontend/react/page.tsx.tmpl",
		"index.ts":                                           "scaffold/frontend/react/index.tsx.tmpl",
	}

	for relPath, tmplPath := range files {
		content, err := g.executeTemplate(relPath, tmplPath, data)
		if err != nil {
			return fmt.Errorf("failed to generate %s: %w", relPath, err)
		}

		structure.Files = append(structure.Files, GeneratedFile{
			Path:    path.Join(moduleDir, relPath),
			Content: content,
		})
	}

	return nil
}

func (g *TemplateGenerator) generateConfiguration(structure *ProjectStructure) error {
	// Generate docker-compose.yml
	dockerComposeFile := GeneratedFile{
		Path:     "docker-compose.yml",
		Template: "docker-compose.yml.tmpl",
		Content:  g.generateDockerCompose(),
	}
	structure.Files = append(structure.Files, dockerComposeFile)

	// Generate Makefile
	makefileFile := GeneratedFile{
		Path:     "Makefile",
		Template: "Makefile.tmpl",
		Content:  g.generateMakefile(),
	}
	structure.Files = append(structure.Files, makefileFile)

	// Generate app config
	configFile := GeneratedFile{
		Path:     "configs/app.yaml",
		Template: "app.yaml.tmpl",
		Content:  g.generateAppConfig(),
	}
	structure.Files = append(structure.Files, configFile)

	return nil
}

func (g *TemplateGenerator) generateBuildScripts(structure *ProjectStructure) error {
	// Generate Dockerfile
	dockerFile := GeneratedFile{
		Path:     "Dockerfile",
		Template: "Dockerfile.tmpl",
		Content:  g.generateDockerfile(),
	}
	structure.Files = append(structure.Files, dockerFile)

	// Generate build script
	buildScript := GeneratedFile{
		Path:       "scripts/build.sh",
		Template:   "build.sh.tmpl",
		Content:    g.generateBuildScript(),
		Executable: true,
	}
	structure.Files = append(structure.Files, buildScript)

	return nil
}

func (g *TemplateGenerator) generateMainTestFile() string {
	content, err := g.executeTemplate("main_test", "scaffold/backend/project/main_test.go.tmpl", nil)
	if err != nil {
		fmt.Printf("⚠️  Error generating main test: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateMigrateMainFile() string {
	driverImport := `_ "modernc.org/sqlite"`
	openDriver := "sqlite"
	openDSN := "app.db"

	switch g.config.Database {
	case "postgres":
		driverImport = `_ "github.com/jackc/pgx/v5/stdlib"`
		openDriver = "pgx"
		openDSN = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
	case "mysql":
		driverImport = `_ "github.com/go-sql-driver/mysql"`
		openDriver = "mysql"
		openDSN = "user:password@tcp(localhost:3306)/dbname?parseTime=true"
	}

	data := map[string]interface{}{
		"DriverImport": driverImport,
		"OpenDriver":   openDriver,
		"OpenDSN":      openDSN,
	}

	content, err := g.executeTemplate("migrate_main", "scaffold/backend/project/migrate_main.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating migrate main: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateCoreProvidersTest() string {
	content, err := g.executeTemplate("core_test", "scaffold/backend/project/core_providers_test.go.tmpl", nil)
	if err != nil {
		fmt.Printf("⚠️  Error generating core test: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateModuleProvidersTestFile(name string) string {
	data := map[string]interface{}{
		"Name": name,
	}

	content, err := g.executeTemplate("module_test", "scaffold/backend/layers/module_providers_test.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating module test: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateRepositoryTestFile(name string) string {
	data := map[string]interface{}{
		"Name":         name,
		"Title":        Capitalize(inflection.Singular(name)),
		"DomainImport": g.moduleImportPath("internal/adapters/http/modules", name, "domain"),
	}

	content, err := g.executeTemplate("repository_test", "scaffold/backend/layers/repository_test.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating repository test: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateServiceTestFile(name string) string {
	data := map[string]interface{}{
		"Name":         name,
		"Title":        Capitalize(inflection.Singular(name)),
		"PluralTitle":  Pluralize(Capitalize(name)),
		"DomainImport": g.moduleImportPath("internal/adapters/http/modules", name, "domain"),
	}

	content, err := g.executeTemplate("service_test", "scaffold/backend/layers/service_test.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating service test: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateHandlerTestFile(name string) string {
	data := map[string]interface{}{
		"Name":         name,
		"Title":        Capitalize(inflection.Singular(name)),
		"PluralTitle":  Pluralize(Capitalize(name)),
		"DomainImport": g.moduleImportPath("internal/adapters/http/modules", name, "domain"),
	}

	content, err := g.executeTemplate("handler_test", "scaffold/backend/layers/handler_test.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating handler test: %v\n", err)
		return ""
	}
	return content
}

// WriteProject writes the generated project to disk
func (g *TemplateGenerator) WriteProject(structure *ProjectStructure) error {
	fmt.Printf("📁 Writing project to: %s\n", structure.RootPath)

	// Create directories
	for _, dir := range structure.Directories {
		dirPath := filepath.Join(structure.RootPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}

	// Write files
	for _, file := range structure.Files {
		filePath := filepath.Join(structure.RootPath, file.Path)

		// Check if file exists and overwrite is disabled
		if !file.Overwrite {
			if _, err := os.Stat(filePath); err == nil {
				fmt.Printf("  ⚠️  Skipping existing file: %s\n", file.Path)
				continue
			}
		}

		// Write file content
		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}

		// Make executable if needed
		if file.Executable {
			if err := os.Chmod(filePath, 0755); err != nil {
				return fmt.Errorf("failed to make file executable %s: %w", filePath, err)
			}
		}

		fmt.Printf("  ✅ Generated: %s\n", file.Path)
	}

	fmt.Printf("🎉 Project generated successfully!\n")
	return nil
}
