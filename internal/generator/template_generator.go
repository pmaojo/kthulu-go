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
	"github.com/pmaojo/kthulu-go/internal/resolver"
	"github.com/pmaojo/kthulu-go/cmd/kthulu/templates"
)

const (
	// Project Templates
	TmplDockerCompose = "docker-compose.yml.tmpl"
	TmplMakefile      = "Makefile.tmpl"
	TmplAppConfig     = "app.yaml.tmpl"
	TmplCIWorkflow    = "github/ci.yaml.tmpl"
	TmplDockerfile    = "Dockerfile.tmpl"
	TmplBuildScript   = "build.sh.tmpl"

	// Go Source Templates
	TmplMainTest      = "scaffold/backend/project/main_test.go.tmpl"
	TmplMigrateMain   = "scaffold/backend/project/migrate_main.go.tmpl"
	TmplCoreProviders = "scaffold/backend/core/providers.go.tmpl"
	TmplConfigGo      = "scaffold/backend/config/config.go.tmpl"
	TmplCoreTest      = "scaffold/backend/project/core_providers_test.go.tmpl"

	// Module Templates
	TmplModule       = "scaffold/backend/module.go.tmpl"
	TmplLayerDomain  = "scaffold/backend/layers/domain.go.tmpl"
	TmplLayerRepo    = "scaffold/backend/layers/repository.go.tmpl"
	TmplLayerService = "scaffold/backend/layers/service.go.tmpl"
	TmplLayerHandler = "scaffold/backend/layers/handler.go.tmpl"
	TmplLayerCore    = "scaffold/backend/layers/core.go.tmpl"
	TmplLayerStore   = "scaffold/backend/layers/store.go.tmpl"
	TmplMigration    = "scaffold/backend/migration.sql.tmpl"

	// Test Templates
	TmplModuleTest  = "scaffold/backend/layers/module_providers_test.go.tmpl"
	TmplRepoTest    = "scaffold/backend/layers/repository_test.go.tmpl"
	TmplServiceTest = "scaffold/backend/layers/service_test.go.tmpl"
	TmplHandlerTest = "scaffold/backend/layers/handler_test.go.tmpl"

	// Frontend Templates (GTH: Go+Templ+HTMX) - The KING of frontend
	TmplGTHBaseLayout   = "scaffold/frontend/gth/layouts/base.templ.tmpl"
	TmplGTHAdminLayout  = "scaffold/frontend/gth/layouts/admin.templ.tmpl"
	TmplGTHTable        = "scaffold/frontend/gth/components/table.templ.tmpl"
	TmplGTHForm         = "scaffold/frontend/gth/components/form.templ.tmpl"
	TmplGTHModal        = "scaffold/frontend/gth/components/modal.templ.tmpl"
	TmplGTHCrudPage     = "scaffold/frontend/gth/pages/crud_page.templ.tmpl"
	TmplGTHDashboard    = "scaffold/frontend/gth/pages/dashboard.templ.tmpl"
	TmplGTHTableRows    = "scaffold/frontend/gth/partials/table_rows.templ.tmpl"
	TmplLayerHandlerGTH = "scaffold/backend/layers/handler_gth.go.tmpl"
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
	TemplateType  string            `json:"template_type"` // server, cli, mcp
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

	// Step 3: Generate structure based on template type
	var errGen error
	switch config.TemplateType {
	case "cli":
		errGen = g.generateCLIStructure(structure)
	case "mcp":
		errGen = g.generateMCPStructure(structure)
	default:
		// Default to server structure
		errGen = g.generateBaseStructure(structure)
	}

	if errGen != nil {
		return nil, fmt.Errorf("failed to generate structure: %w", errGen)
	}

	// Step 4: Generate module files (only for server templates)
	if config.TemplateType == "server" || config.TemplateType == "" {
		for _, module := range plan.InstallOrder {
			if err := g.generateModuleFiles(module, structure); err != nil {
				return nil, fmt.Errorf("failed to generate module '%s': %w", module, err)
			}
		}
	}

	// Step 5: Generate GTH (Go+Templ+HTMX) frontend if not disabled
	if config.Frontend != "none" {
		if err := g.generateGTHFrontend(structure); err != nil {
			return nil, fmt.Errorf("failed to generate GTH frontend: %w", err)
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
		"internal/infrastructure/server",
		"internal/infrastructure/static",
		"internal/infrastructure/static/dist",
		"internal/infrastructure/config",
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
		"internal/infrastructure/server/server.go":         "scaffold/backend/internal/infrastructure/server/server.go.tmpl",
		"internal/infrastructure/static/fs.go":             "scaffold/backend/internal/infrastructure/static/fs.go.tmpl",
		"internal/infrastructure/config/config.go":         "scaffold/backend/internal/infrastructure/config/config.go.tmpl",
	}

	// Generate placeholder static file to prevent go:embed error
	structure.Files = append(structure.Files, GeneratedFile{
		Path:    "internal/infrastructure/static/dist/index.html",
		Content: "<h1>Welcome to Kthulu Omega Engine</h1>",
	})

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
		fmt.Sprintf("%s/core", moduleBase),
		fmt.Sprintf("%s/store", moduleBase),
		fmt.Sprintf("%s/api", moduleBase),
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
		for {
			collision := false
			timestampPrefix := filepath.Join("migrations", timestamp+"_")
			for _, f := range structure.Files {
				if strings.HasPrefix(f.Path, timestampPrefix) {
					collision = true
					break
				}
			}
			if !collision {
				break
			}
			// If collision, wait 1s and try again
			time.Sleep(1 * time.Second)
			timestamp = time.Now().Format("20060102150405")
		}

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
	addDep(fmt.Sprintf("gorm.io/driver/%s v1.5.4", g.config.Database))
	addDep("gorm.io/driver/sqlite v1.5.4")
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

	// Generate domain imports for AutoMigrate
	var domainImports []string
	var autoMigrateModels []string
	plan, _ := g.resolver.ResolveDependencies(g.config.Features)
	relPath := g.getModuleRelPath()
	for _, module := range plan.RequiredModules {
		coreImport := g.moduleImportPath(relPath, module, "core")
		domainImports = append(domainImports, fmt.Sprintf("\t%sCore \"%s\"", module, coreImport))
		autoMigrateModels = append(autoMigrateModels, fmt.Sprintf("&%sCore.%s{}", module, Capitalize(module)))
	}

	autoMigrateCall := ""
	if len(autoMigrateModels) > 0 {
		autoMigrateCall = fmt.Sprintf(`
    // Auto-migrate all domain models
    if err := db.AutoMigrate(%s); err != nil {
        return nil, fmt.Errorf("auto-migrate failed: %%w", err)
    }
    return db, nil`, strings.Join(autoMigrateModels, ", "))
	}

	data := struct {
		Database        string
		DBName          string
		DomainImports   []string
		AutoMigrateCall string
	}{
		Database:        dbDriver,
		DBName:          dbName,
		DomainImports:   domainImports,
		AutoMigrateCall: autoMigrateCall,
	}

	content, err := g.executeTemplate("coreProviders", "scaffold/backend/core/providers.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️ Failed to generate core providers: %v\n", err)
		return ""
	}
	return content
}

// Helper methods for code generation
func (g *TemplateGenerator) generateModuleImports() string {
	var imports []string
	relPath := g.getModuleRelPath()
	// Use resolved dependencies, not just initial features
	plan, _ := g.resolver.ResolveDependencies(g.config.Features)
	for _, module := range plan.RequiredModules {
		moduleBase := g.moduleImportPath(relPath, module)
		coreImport := g.moduleImportPath(relPath, module, "core")
		apiImport := g.moduleImportPath(relPath, module, "api")
		imports = append(imports, fmt.Sprintf(` "%s"`, moduleBase))
		imports = append(imports, fmt.Sprintf(` %sCore "%s"`, module, coreImport))
		imports = append(imports, fmt.Sprintf(` %sAPI "%s"`, module, apiImport))
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
		routes = append(routes, fmt.Sprintf(`	%sHandler := %sAPI.New%sHandler(%sService)`, module, module, Capitalize(module), module))
		routes = append(routes, fmt.Sprintf(`	%sHandler.RegisterRoutes(apiRouter)`, module))
	}
	return strings.Join(routes, "\n")
}

func (g *TemplateGenerator) generateInvokeParams() string {
	var params []string
	plan, _ := g.resolver.ResolveDependencies(g.config.Features)
	for _, module := range plan.RequiredModules {
		params = append(params, fmt.Sprintf(`%sService %sCore.%sService`, module, module, Capitalize(module)))
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

	content, err := g.executeTemplate("module", TmplModule, data)
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
		"CoreImport":   g.moduleImportPath(relPath, name, "core"),
	}

	content, err := g.executeTemplate("repository", TmplLayerRepo, data)
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
		"CoreImport":   g.moduleImportPath(relPath, name, "core"),
	}

	content, err := g.executeTemplate("service", TmplLayerService, data)
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

	content, err := g.executeTemplate("handler", TmplLayerHandler, data)
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
		"CoreImport":    g.moduleImportPath(moduleRelPath, moduleName, "core"),
	}

	files := make(map[string]string)
	
	// Core layers
	layers := map[string]string{
		"module.go":                             "scaffold/backend/module.go.tmpl",
		"core/" + moduleName + ".go":            TmplLayerCore,
		"store/" + moduleName + "_store.go":     TmplLayerStore,
		"core/" + moduleName + "_service.go":    TmplLayerService,
		"api/" + moduleName + "_handler.go":     TmplLayerHandler,
	}

	for relPath, tmplPath := range layers {
		content, err := g.executeTemplate(relPath, tmplPath, data)
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate %s: %w", relPath, err)
		}
		files[relPath] = content
	}

	// Migration
	migrationContent, err := g.executeTemplate("migration", TmplMigration, data)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to generate migration: %v\n", err)
	}

	return files, migrationContent, nil
}
// generateGTHFrontend generates GTH (Go+Templ+HTMX) frontend structure
func (g *TemplateGenerator) generateGTHFrontend(structure *ProjectStructure) error {
	fmt.Println("  🎨 Setting up GTH (Go+Templ+HTMX) frontend...")

	// Create views directory structure inside internal/
	dirs := []string{
		"internal/views",
		"internal/views/layouts",
		"internal/views/components",
		"internal/views/pages",
		"internal/views/partials",
	}
	structure.Directories = append(structure.Directories, dirs...)

	// Generate base layouts
	layoutData := map[string]interface{}{
		"ProjectName":   g.config.ProjectName,
		"ProjectModule": g.modulePath(),
		"Features":      g.config.Features,
	}

	// Base layout
	baseContent, err := g.executeTemplate("gth_base", TmplGTHBaseLayout, layoutData)
	if err != nil {
		return fmt.Errorf("failed to generate base layout: %w", err)
	}
	structure.Files = append(structure.Files, GeneratedFile{
		Path:    "internal/views/layouts/base.templ",
		Content: baseContent,
	})

	// Admin layout
	adminContent, err := g.executeTemplate("gth_admin", TmplGTHAdminLayout, layoutData)
	if err != nil {
		return fmt.Errorf("failed to generate admin layout: %w", err)
	}
	structure.Files = append(structure.Files, GeneratedFile{
		Path:    "internal/views/layouts/admin.templ",
		Content: adminContent,
	})

	// Modal component (shared)
	modalContent, err := g.executeTemplate("gth_modal", TmplGTHModal, layoutData)
	if err != nil {
		return fmt.Errorf("failed to generate modal component: %w", err)
	}
	structure.Files = append(structure.Files, GeneratedFile{
		Path:    "internal/views/components/modal.templ",
		Content: modalContent,
	})

	// Dashboard page
	dashContent, err := g.executeTemplate("gth_dashboard", TmplGTHDashboard, layoutData)
	if err != nil {
		return fmt.Errorf("failed to generate dashboard: %w", err)
	}
	structure.Files = append(structure.Files, GeneratedFile{
		Path:    "internal/views/pages/dashboard.templ",
		Content: dashContent,
	})

	// Generate module-specific views
	for _, feature := range g.config.Features {
		if feature == "auth" || feature == "users" {
			continue
		}

		fields := g.config.ModuleFields[feature]
		if len(fields) == 0 {
			fields = []string{"name:string"}
		}

		if err := g.GenerateGTHModule(feature, fields, structure); err != nil {
			return fmt.Errorf("failed to generate GTH module for %s: %w", feature, err)
		}
	}

	return nil
}

// GenerateGTHModule generates GTH views for a specific module
func (g *TemplateGenerator) GenerateGTHModule(moduleName string, fields []string, structure *ProjectStructure) error {
	fmt.Printf("  🎨 Generating GTH views for: %s\n", moduleName)

	backendFields := ParseBackendFields(fields)
	relPath := g.getModuleRelPath()

	data := map[string]interface{}{
		"Name":          moduleName,
		"Title":         Capitalize(moduleName),
		"PluralTitle":   Pluralize(Capitalize(moduleName)),
		"Fields":        backendFields,
		"ProjectModule": g.modulePath(),
		"RoutePrefix":   ToKebabCase(moduleName),
		"CoreImport":    g.moduleImportPath(relPath, moduleName, "core"),
	}

	// Generate module-specific templ files
	templates := map[string]string{
		fmt.Sprintf("internal/views/components/%s_table.templ", moduleName):    TmplGTHTable,
		fmt.Sprintf("internal/views/components/%s_form.templ", moduleName):     TmplGTHForm,
		fmt.Sprintf("internal/views/pages/%s_page.templ", moduleName):          TmplGTHCrudPage,
		fmt.Sprintf("internal/views/partials/%s_table_rows.templ", moduleName): TmplGTHTableRows,
	}

	for filePath, tmplPath := range templates {
		content, err := g.executeTemplate(filePath, tmplPath, data)
		if err != nil {
			return fmt.Errorf("failed to generate %s: %w", filePath, err)
		}
		structure.Files = append(structure.Files, GeneratedFile{
			Path:    filePath,
			Content: content,
		})
	}

	return nil
}

func (g *TemplateGenerator) generateConfiguration(structure *ProjectStructure) error {
	// Generate docker-compose.yml
	dockerComposeFile := GeneratedFile{
		Path:     "docker-compose.yml",
		Template: TmplDockerCompose,
		Content:  g.generateDockerCompose(),
	}
	structure.Files = append(structure.Files, dockerComposeFile)

	// Generate Makefile
	makefileFile := GeneratedFile{
		Path:     "Makefile",
		Template: TmplMakefile,
		Content:  g.generateMakefile(),
	}
	structure.Files = append(structure.Files, makefileFile)

	// Generate app config
	configFile := GeneratedFile{
		Path:     "configs/app.yaml",
		Template: TmplAppConfig,
		Content:  g.generateAppConfig(),
	}
	structure.Files = append(structure.Files, configFile)

	// Generate internal/config/config.go
	configGoFile := GeneratedFile{
		Path:     "internal/config/config.go",
		Template: TmplConfigGo,
		Content:  g.generateConfigGo(),
	}
	structure.Files = append(structure.Files, configGoFile)

	// Generate GitHub CI Workflow
	ciFile := GeneratedFile{
		Path:     ".github/workflows/ci.yaml",
		Template: TmplCIWorkflow,
		Content:  "", // Loaded dynamically
	}
	if content, err := g.executeTemplate("ciWorkflow", TmplCIWorkflow, nil); err == nil {
		ciFile.Content = content
		structure.Files = append(structure.Files, ciFile)
	}

	return nil
}

func (g *TemplateGenerator) generateBuildScripts(structure *ProjectStructure) error {
	// Generate Dockerfile
	dockerFile := GeneratedFile{
		Path:     "Dockerfile",
		Template: TmplDockerfile,
		Content:  g.generateDockerfile(),
	}
		structure.Files = append(structure.Files, dockerFile)

	// Generate build script
	buildScript := GeneratedFile{
		Path:       "scripts/build.sh",
		Template:   TmplBuildScript,
		Content:    g.generateBuildScript(),
		Executable: true,
	}
	structure.Files = append(structure.Files, buildScript)

	return nil
}

func (g *TemplateGenerator) generateMainTestFile() string {
	content, err := g.executeTemplate("main_test", TmplMainTest, nil)
	if err != nil {
		fmt.Printf("⚠️  Error generating main test: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateMigrateMainFile() string {
	driverImport := `_ "github.com/mattn/go-sqlite3"`
	openDriver := "sqlite3"
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
		"ModulePath":   g.modulePath(),
	}

	content, err := g.executeTemplate("migrate_main", TmplMigrateMain, data)
	if err != nil {
		fmt.Printf("⚠️  Error generating migrate main: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateConfigGo() string {
	data := map[string]interface{}{
		"Database":    g.config.Database,
		"ProjectName": g.config.ProjectName,
	}
	content, err := g.executeTemplate("config_go", TmplConfigGo, data)
	if err != nil {
		fmt.Printf("⚠️  Error generating config.go: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateCoreProvidersTest() string {
	content, err := g.executeTemplate("core_test", TmplCoreTest, nil)
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

	content, err := g.executeTemplate("module_test", TmplModuleTest, data)
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

	content, err := g.executeTemplate("repository_test", TmplRepoTest, data)
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
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
		}
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

// ResultRegisterFrontendNavigation registers a module in the frontend navigation config
func (g *TemplateGenerator) ResultRegisterFrontendNavigation(title, name, rootPath string) error {
	configPath := filepath.Join(rootPath, "frontend/src/config/navigation.ts")
	
	// Read existing file
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Config doesn't exist, skip
		}
		return err
	}
	
	text := string(content)
	
	// Check if already registered
	if strings.Contains(text, `path: '/`+name+`'`) {
		return nil
	}
	
	// Prepare new item
	newItem := fmt.Sprintf(`  {
    title: '%s',
    path: '/%s',
    icon: LayoutDashboard,
    category: 'Modules'
  },
`, title, name)

	// Insert before the end of the array
	lastBracket := strings.LastIndex(text, "];")
	if lastBracket == -1 {
		return fmt.Errorf("could not find closing bracket in navigation.ts")
	}
	
	newText := text[:lastBracket] + newItem + text[lastBracket:]
	
	return os.WriteFile(configPath, []byte(newText), 0644)
}

func (g *TemplateGenerator) generateCLIStructure(structure *ProjectStructure) error {
	structure.Directories = append(structure.Directories, "cmd/"+g.config.ProjectName, "internal/cli")

	// main.go
	structure.Files = append(structure.Files, GeneratedFile{
		Path:     "cmd/" + g.config.ProjectName + "/main.go",
		Template: "cli_main",
		Content:  g.generateCLIMain(),
	})

	// root.go
	structure.Files = append(structure.Files, GeneratedFile{
		Path:     "internal/cli/root.go",
		Template: "cli_root",
		Content:  g.generateCLIRoot(),
	})

	// go.mod
	structure.Files = append(structure.Files, GeneratedFile{
		Path:     "go.mod",
		Template: "cli_gomod",
		Content:  g.generateCLIGoMod(),
	})

	// README.md
	structure.Files = append(structure.Files, GeneratedFile{
		Path:     "README.md",
		Template: "readme",
		Content:  g.generateReadme(),
	})

	return nil
}

func (g *TemplateGenerator) generateMCPStructure(structure *ProjectStructure) error {
	structure.Directories = append(structure.Directories, "cmd/"+g.config.ProjectName, "internal/tools")

	// main.go
	structure.Files = append(structure.Files, GeneratedFile{
		Path:     "cmd/" + g.config.ProjectName + "/main.go",
		Template: "mcp_main",
		Content:  g.generateMCPMain(),
	})

	// tools.go
	structure.Files = append(structure.Files, GeneratedFile{
		Path:     "internal/tools/tools.go",
		Template: "mcp_tools",
		Content:  g.generateMCPTools(),
	})

	// go.mod
	structure.Files = append(structure.Files, GeneratedFile{
		Path:     "go.mod",
		Template: "mcp_gomod",
		Content:  g.generateMCPGoMod(),
	})

	// README.md
	structure.Files = append(structure.Files, GeneratedFile{
		Path:     "README.md",
		Template: "readme",
		Content:  g.generateReadme(),
	})

	return nil
}

func (g *TemplateGenerator) generateCLIMain() string {
	data := map[string]interface{}{
		"ModulePath": g.modulePath(),
	}
	content, err := g.executeTemplate("cli_main", "scaffold/cli/main.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating CLI main.go: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateCLIRoot() string {
	data := map[string]interface{}{
		"ProjectName": g.config.ProjectName,
	}
	content, err := g.executeTemplate("cli_root", "scaffold/cli/root.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating CLI root.go: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateCLIGoMod() string {
	data := map[string]interface{}{
		"ModulePath": g.modulePath(),
		"GoVersion":  "1.24",
	}
	content, err := g.executeTemplate("cli_gomod", "scaffold/cli/go.mod.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating CLI go.mod: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateMCPMain() string {
	data := map[string]interface{}{
		"ProjectName": g.config.ProjectName,
		"ModulePath":  g.modulePath(),
	}
	content, err := g.executeTemplate("mcp_main", "scaffold/mcp/main.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating MCP main.go: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateMCPTools() string {
	data := map[string]interface{}{
		"ProjectName": g.config.ProjectName,
	}
	content, err := g.executeTemplate("mcp_tools", "scaffold/mcp/tools.go.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating MCP tools.go: %v\n", err)
		return ""
	}
	return content
}

func (g *TemplateGenerator) generateMCPGoMod() string {
	data := map[string]interface{}{
		"ModulePath": g.modulePath(),
		"GoVersion":  "1.24",
	}
	content, err := g.executeTemplate("mcp_gomod", "scaffold/mcp/go.mod.tmpl", data)
	if err != nil {
		fmt.Printf("⚠️  Error generating MCP go.mod: %v\n", err)
		return ""
	}
	return content
}
