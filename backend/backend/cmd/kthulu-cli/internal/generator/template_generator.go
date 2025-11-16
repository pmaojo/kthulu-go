package generator

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/resolver"
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
	Database      string            `json:"database"`      // sqlite, postgres, mysql
	Auth          string            `json:"auth"`          // jwt, oauth, both
	Features      []string          `json:"features"`      // modules to include
	Enterprise    bool              `json:"enterprise"`    // enterprise features
	Observability bool              `json:"observability"` // monitoring
	CustomValues  map[string]string `json:"custom_values"` // custom template values
}

// modulePath returns the module import path for the generated project.
func (g *TemplateGenerator) modulePath() string {
	if g.config == nil {
		return ""
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

// NewTemplateGenerator creates a new template generator
func NewTemplateGenerator(resolver *resolver.DependencyResolver) *TemplateGenerator {
	return &TemplateGenerator{
		resolver:  resolver,
		templates: make(map[string]*template.Template),
		config:    &GeneratorConfig{},
	}
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
	if config.Frontend != "none" {
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
		"cmd/cli",
		"cmd/migrate",
		"internal/core",
		"internal/adapters/http",
		"internal/adapters/http/modules",
		"internal/adapters/cli",
		"internal/adapters/mcp",
		"internal/domain",
		"internal/domain/repository",
		"internal/usecase",
		"internal/infrastructure",
		"pkg/utils",
		"pkg/errors",
		"configs",
		"migrations",
		"scripts",
		"docs",
		"deployments",
		"test",
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

	return nil
}

// generateModuleFiles generates files for a specific module
func (g *TemplateGenerator) generateModuleFiles(moduleName string, structure *ProjectStructure) error {
	fmt.Printf("  📦 Generating module: %s\n", moduleName)

	// Get module information
	moduleInfo, err := g.resolver.GetModuleInfo(moduleName)
	if err != nil {
		// Generate basic module structure if not found
		moduleInfo = &resolver.ModuleInfo{
			Name:        moduleName,
			Category:    "Custom",
			Description: fmt.Sprintf("Custom %s module", moduleName),
		}
	}

	// Generate module directory structure
	moduleBase := fmt.Sprintf("internal/adapters/http/modules/%s", moduleName)
	moduleDirs := []string{
		moduleBase,
		fmt.Sprintf("%s/domain", moduleBase),
		fmt.Sprintf("%s/repository", moduleBase),
		fmt.Sprintf("%s/service", moduleBase),
		fmt.Sprintf("%s/handlers", moduleBase),
		fmt.Sprintf("%s/dto", moduleBase),
	}
	structure.Directories = append(structure.Directories, moduleDirs...)

	// Generate module files
	moduleFiles := []GeneratedFile{
		{
			Path:     fmt.Sprintf("%s/module.go", moduleBase),
			Template: "module.go.tmpl",
			Content:  g.generateModuleFile(moduleName, moduleInfo),
		},
		{
			Path:     fmt.Sprintf("%s/domain/%s.go", moduleBase, moduleName),
			Template: "domain.go.tmpl",
			Content:  g.generateDomainFileFixed(moduleName, moduleInfo),
		},
		{
			Path:     fmt.Sprintf("%s/repository/%s_repository.go", moduleBase, moduleName),
			Template: "repository.go.tmpl",
			Content:  g.generateRepositoryFileFixed(moduleName, moduleInfo),
		},
		{
			Path:     fmt.Sprintf("%s/service/%s_service.go", moduleBase, moduleName),
			Template: "service.go.tmpl",
			Content:  g.generateServiceFileFixed(moduleName, moduleInfo),
		},
		{
			Path:     fmt.Sprintf("%s/handlers/%s_handler.go", moduleBase, moduleName),
			Template: "handler.go.tmpl",
			Content:  g.generateHandlerFile(moduleName, moduleInfo),
		},
	}

	structure.Files = append(structure.Files, moduleFiles...)

	// Generate module tests
	testFiles := []GeneratedFile{
		{
			Path:    fmt.Sprintf("%s/module_test.go", moduleBase),
			Content: g.generateModuleProvidersTestFile(moduleName),
		},
		{
			Path:    fmt.Sprintf("%s/repository/%s_repository_test.go", moduleBase, moduleName),
			Content: g.generateRepositoryTestFile(moduleName),
		},
		{
			Path:    fmt.Sprintf("%s/service/%s_service_test.go", moduleBase, moduleName),
			Content: g.generateServiceTestFile(moduleName),
		},
		{
			Path:    fmt.Sprintf("%s/handlers/%s_handler_test.go", moduleBase, moduleName),
			Content: g.generateHandlerTestFile(moduleName),
		},
	}

	structure.Files = append(structure.Files, testFiles...)

	return nil
}

// generateMainFile generates the main.go file
func (g *TemplateGenerator) generateMainFile() string {
	coreImport := g.moduleImportPath("internal/core")
	return fmt.Sprintf(`// @kthulu:project:%s
// @kthulu:generated:true
// @kthulu:features:%s
package main

import (
"context"
"log"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/gorilla/mux"
"go.uber.org/fx"

"%s"
%s
)

type httpServer interface {
Start() error
Shutdown(context.Context) error
}

type realHTTPServer struct {
server *http.Server
}

func newHTTPServer(handler http.Handler) httpServer {
return &realHTTPServer{
server: &http.Server{
Addr:    ":8080",
Handler: handler,
},
}
}

func (s *realHTTPServer) Start() error {
if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
return err
}
return nil
}

func (s *realHTTPServer) Shutdown(ctx context.Context) error {
return s.server.Shutdown(ctx)
}

type noopHTTPServer struct{}

func (n *noopHTTPServer) Start() error {
return nil
}

func (n *noopHTTPServer) Shutdown(context.Context) error {
return nil
}

var serverBuilder = func(handler http.Handler) httpServer {
if os.Getenv("KTHULU_TEST_MODE") == "1" {
return &noopHTTPServer{}
}
return newHTTPServer(handler)
}

func main() {
if err := runApplication(context.Background(), serverBuilder); err != nil {
log.Fatal("Failed to start application:", err)
}
}

func runApplication(ctx context.Context, builder func(http.Handler) httpServer) error {
ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
defer stop()

app := fx.New(
// Core providers
core.CoreRepositoryProviders(),

// Module providers
%s

fx.Invoke(func(lc fx.Lifecycle, %s) {
router := setupRoutes()
apiRouter := router.PathPrefix("/api/v1").Subrouter()

%s

server := builder(router)

lc.Append(fx.Hook{
OnStart: func(context.Context) error {
go func() {
if err := server.Start(); err != nil {
log.Println("server error:", err)
}
}()
return nil
},
OnStop: func(ctx context.Context) error {
return server.Shutdown(ctx)
},
})
}),
)

if err := app.Start(ctx); err != nil {
return err
}

<-ctx.Done()

shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

return app.Stop(shutdownCtx)
}

func setupRoutes() *mux.Router {
router := mux.NewRouter()

router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
_, _ = w.Write([]byte("OK"))
})

return router
}
`, g.config.ProjectName,
		strings.Join(g.config.Features, ","),
		coreImport,
		g.generateModuleImports(),
		g.generateModuleProviders(),
		g.generateInvokeParams(),
		g.generateModuleRoutes())
}

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

	if extra := strings.Split(strings.TrimSpace(g.generateDependencies()), "\n"); len(extra) > 0 {
		for _, dep := range extra {
			addDep(strings.TrimSpace(dep))
		}
	}

	return fmt.Sprintf(`module %s

go 1.21

require (
%s
)
`, modulePath, strings.Join(deps, "\n"))
}

// generateReadme generates the README.md file
func (g *TemplateGenerator) generateReadme() string {
	readme := fmt.Sprintf(`# %s

A Kthulu-powered enterprise application with the following features:

## Features
%s

## Architecture
- **Framework**: Kthulu Enterprise
- **Database**: %s
- **Frontend**: %s
- **Authentication**: %s

## Quick Start

`+"```bash"+`
# Install dependencies
go mod download

# Run migrations
go run cmd/migrate/main.go

# Start development server
go run cmd/server/main.go
`+"```"+`

## Development

### Adding New Modules
`+"```bash"+`
kthulu add module <module-name>
`+"```"+`

### Running Tests
`+"```bash"+`
go test ./...
`+"```"+`

### Building for Production
`+"```bash"+`
go build -o bin/server cmd/server/main.go
`+"```"+`

## Generated by Kthulu CLI
This project was generated using Kthulu CLI with intelligent dependency resolution.

Features: %s
Dependencies resolved: %s
`, g.config.ProjectName,
		g.generateFeatureList(),
		g.config.Database,
		g.config.Frontend,
		g.config.Auth,
		strings.Join(g.config.Features, ", "),
		fmt.Sprintf("%d modules", len(g.config.Features)))

	return readme
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
		"\"gorm.io/driver/sqlite\"",
	}

	var connectionBuilder strings.Builder
	var driverImport string
	switch dbDriver {
	case "postgres":
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
		driverImport = "\"gorm.io/driver/sqlite\""
		imports = append(imports, "\"path/filepath\"")
		connectionBuilder.WriteString(fmt.Sprintf("\t\tdbPath := fmt.Sprintf(\"%%s\", getEnv(\"SQLITE_PATH\", \"data/%s.db\"))\n", dbName))
		connectionBuilder.WriteString("\t\tif err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {\n")
		connectionBuilder.WriteString("\t\t\treturn nil, err\n")
		connectionBuilder.WriteString("\t\t}\n")
		connectionBuilder.WriteString("\t\tlog.Printf(\"Using SQLite database at %s\", dbPath)\n")
		connectionBuilder.WriteString("\t\treturn gorm.Open(sqlite.Open(dbPath), &gorm.Config{})\n")
	}
	imports = append(imports, driverImport)

	var importLines []string
	for _, imp := range imports {
		importLines = append(importLines, "\t"+imp)
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
return gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
}
%s
}

func getEnv(key, fallback string) string {
        if value := os.Getenv(key); value != "" {
                return value
        }
        return fallback
}
`, strings.Join(importLines, "\n"), connectionBuilder.String())
}

// Helper methods for code generation
func (g *TemplateGenerator) generateModuleImports() string {
	var imports []string
	// Use resolved dependencies, not just initial features
	plan, _ := g.resolver.ResolveDependencies(g.config.Features)
	for _, module := range plan.RequiredModules {
		moduleBase := g.moduleImportPath("internal/adapters/http/modules", module)
		domainImport := g.moduleImportPath("internal/adapters/http/modules", module, "domain")
		handlersImport := g.moduleImportPath("internal/adapters/http/modules", module, "handlers")
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

// Additional generation methods (simplified for brevity)
func (g *TemplateGenerator) generateModuleFile(name string, info *resolver.ModuleInfo) string {
	capName := Capitalize(name)
	repositoryImport := g.moduleImportPath("internal/adapters/http/modules", name, "repository")
	serviceImport := g.moduleImportPath("internal/adapters/http/modules", name, "service")
	handlersImport := g.moduleImportPath("internal/adapters/http/modules", name, "handlers")

	return fmt.Sprintf(`// @kthulu:module:%s
// @kthulu:category:%s
package %s

import (
"go.uber.org/fx"

"%s"
"%s"
"%s"
)

// Providers returns the Fx providers for the %s module
func Providers() fx.Option {
return fx.Options(
fx.Provide(
repository.New%sRepository,
service.New%sService,
handlers.New%sHandler,
),
)
}
`, name, info.Category, name,
		repositoryImport,
		serviceImport,
		handlersImport,
		name,
		capName,
		capName,
		capName)
}

func (g *TemplateGenerator) generateDomainFile(name string, info *resolver.ModuleInfo) string {
	capName := Capitalize(name)
	pluralName := Pluralize(capName)
	return fmt.Sprintf(`// @kthulu:domain:%s
package domain

import "time"

// %s represents a %s entity
type %s struct {
	ID        uint      `+"`"+`json:"id" gorm:"primaryKey"`+"`"+`
	CreatedAt time.Time `+"`"+`json:"created_at"`+"`"+`
	UpdatedAt time.Time `+"`"+`json:"updated_at"`+"`"+`
	
	// Add your fields here
}

// %sRepository defines the repository interface
type %sRepository interface {
	Create(entity *%s) error
	GetByID(id uint) (*%s, error)
	Update(entity *%s) error
	Delete(id uint) error
	List() ([]*%s, error)
}

// %sService defines the service interface  
type %sService interface {
	Create%s(entity *%s) error
	Get%sByID(id uint) (*%s, error)
	Update%s(entity *%s) error
	Delete%s(id uint) error
	List%s() ([]*%s, error)
}
`, name, capName, name, capName,
		capName, capName, capName, capName, capName, capName,
		capName, capName, capName, capName, capName, capName,
		capName, capName, pluralName, capName, capName)
}

func (g *TemplateGenerator) generateRepositoryFile(name string, info *resolver.ModuleInfo) string {
	capName := Capitalize(name)
	return fmt.Sprintf(`// @kthulu:repository:%s
package repository

import (
	"gorm.io/gorm"
	"%s"
)

type %sRepository struct {
	db *gorm.DB
}

func New%sRepository(db *gorm.DB) domain.%sRepository {
	return &%sRepository{db: db}
}

func (r *%sRepository) Create(entity *domain.%s) error {
	return r.db.Create(entity).Error
}

func (r *%sRepository) GetByID(id uint) (*domain.%s, error) {
	var entity domain.%s
	err := r.db.First(&entity, id).Error
	return &entity, err
}

func (r *%sRepository) Update(entity *domain.%s) error {
	return r.db.Save(entity).Error
}

func (r *%sRepository) Delete(id uint) error {
	return r.db.Delete(&domain.%s{}, id).Error
}

func (r *%sRepository) List() ([]*domain.%s, error) {
	var entities []*domain.%s
	err := r.db.Find(&entities).Error
	return entities, err
}
`, name, g.moduleImportPath("internal/adapters/http/modules", name, "domain"), capName, capName, capName,
		capName, capName, capName, capName, capName, capName,
		capName, capName, capName, capName, capName, capName,
		capName)
}

func (g *TemplateGenerator) generateServiceFile(name string, info *resolver.ModuleInfo) string {
	capName := Capitalize(name)
	pluralName := Pluralize(capName)
	return fmt.Sprintf(`// @kthulu:service:%s
package service

import (
	"%s"
)

type %sService struct {
	repo domain.%sRepository
}

func New%sService(repo domain.%sRepository) domain.%sService {
	return &%sService{repo: repo}
}

func (s *%sService) Create%s(entity *domain.%s) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *%sService) Get%sByID(id uint) (*domain.%s, error) {
	return s.repo.GetByID(id)
}

func (s *%sService) Update%s(entity *domain.%s) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *%sService) Delete%s(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *%sService) List%s() ([]*domain.%s, error) {
	return s.repo.List()
}
`, name, g.moduleImportPath("internal/adapters/http/modules", name, "domain"), capName, capName, capName,
		capName, capName, capName, capName, capName, capName,
		capName, capName, capName, capName, capName, capName,
		capName, capName, capName, pluralName, capName)
}

func (g *TemplateGenerator) generateHandlerFile(name string, info *resolver.ModuleInfo) string {
	capName := Capitalize(name)
	pluralName := Pluralize(capName)
	return fmt.Sprintf(`// @kthulu:handler:%s
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"%s"
)

type %sHandler struct {
	service domain.%sService
}

func New%sHandler(service domain.%sService) *%sHandler {
	return &%sHandler{service: service}
}

// RegisterRoutes registers all routes for %s
func (h *%sHandler) RegisterRoutes(router *mux.Router) {
	sub := router.PathPrefix("/%s").Subrouter()
	sub.HandleFunc("", h.List).Methods("GET")
	sub.HandleFunc("", h.Create).Methods("POST")
	sub.HandleFunc("/{id}", h.GetByID).Methods("GET")
	sub.HandleFunc("/{id}", h.Update).Methods("PUT")
	sub.HandleFunc("/{id}", h.Delete).Methods("DELETE")
}

func (h *%sHandler) Create(w http.ResponseWriter, r *http.Request) {
	var entity domain.%s
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := h.service.Create%s(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

func (h *%sHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	entity, err := h.service.Get%sByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

func (h *%sHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var entity domain.%s
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	entity.ID = uint(id)
	if err := h.service.Update%s(&entity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

func (h *%sHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	if err := h.service.Delete%s(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	entities, err := h.service.List%s()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
`, name, g.moduleImportPath("internal/adapters/http/modules", name, "domain"), capName, capName, capName,
		capName, capName, capName, name, capName, name,
		capName, capName, capName, capName, capName,
		capName, capName, capName, capName, capName, capName, pluralName)
}

// Additional methods for frontend, configuration, and build scripts
func (g *TemplateGenerator) generateFrontend(structure *ProjectStructure) error {
	// Implementation for frontend generation based on config.Frontend
	// React, Templ, Fyne, etc.
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
	return `// @kthulu:test:cmd:server
package main

import (
"context"
"net/http"
"net/http/httptest"
"testing"
"time"
)

type testHTTPServer struct {
started  bool
shutdown bool
}

func (s *testHTTPServer) Start() error {
s.started = true
return nil
}

func (s *testHTTPServer) Shutdown(context.Context) error {
s.shutdown = true
return nil
}

func TestRunApplicationLifecycle(t *testing.T) {
t.Setenv("KTHULU_TEST_MODE", "1")
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

srv := &testHTTPServer{}
errCh := make(chan error, 1)

go func() {
errCh <- runApplication(ctx, func(http.Handler) httpServer {
return srv
})
}()

time.Sleep(20 * time.Millisecond)
cancel()

select {
case err := <-errCh:
if err != nil {
t.Fatalf("runApplication returned error: %%v", err)
}
case <-time.After(time.Second):
t.Fatal("timeout waiting for application shutdown")
}

if !srv.started || !srv.shutdown {
t.Fatalf("server lifecycle not executed: started=%%v shutdown=%%v", srv.started, srv.shutdown)
}
}

func TestSetupRoutesHealth(t *testing.T) {
handler := setupRoutes()
req := httptest.NewRequest(http.MethodGet, "/health", nil)
rr := httptest.NewRecorder()
handler.ServeHTTP(rr, req)

if rr.Code != http.StatusOK {
t.Fatalf("expected status 200 got %d", rr.Code)
}
if body := rr.Body.String(); body != "OK" {
t.Fatalf("expected OK body got %s", body)
}
}
`
}

func (g *TemplateGenerator) generateCoreProvidersTest() string {
	return `// @kthulu:test:core
package core

import "testing"

func TestGetEnv(t *testing.T) {
t.Setenv("KTHULU_SAMPLE", "value")
if got := getEnv("KTHULU_SAMPLE", "fallback"); got != "value" {
t.Fatalf("expected value got %s", got)
}
if got := getEnv("MISSING", "fallback"); got != "fallback" {
t.Fatalf("expected fallback got %s", got)
}
}

func TestNewDatabaseTestMode(t *testing.T) {
t.Setenv("KTHULU_TEST_MODE", "1")
db, err := NewDatabase()
if err != nil {
t.Fatalf("expected sqlite database, got error: %%v", err)
}
if db == nil {
t.Fatal("expected database instance")
}
}
`
}

func (g *TemplateGenerator) generateModuleProvidersTestFile(name string) string {
	return fmt.Sprintf(`// @kthulu:test:module:%[1]s
package %[1]s

import "testing"

func TestProviders(t *testing.T) {
if Providers() == nil {
t.Fatal("expected providers option")
}
}
`, name)
}

func (g *TemplateGenerator) generateRepositoryTestFile(name string) string {
	capName := Capitalize(name)
	domainImport := g.moduleImportPath("internal/adapters/http/modules", name, "domain")
	return fmt.Sprintf(`// @kthulu:test:repository:%[1]s
package repository

import (
"testing"

"gorm.io/driver/sqlite"
"gorm.io/gorm"

"%[2]s"
)

func Test%[3]sRepositoryCRUD(t *testing.T) {
db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
if err != nil {
t.Fatalf("failed to open sqlite: %%v", err)
}
if err := db.AutoMigrate(&domain.%[3]s{}); err != nil {
t.Fatalf("failed to migrate: %%v", err)
}
repo := New%[3]sRepository(db)
entity := &domain.%[3]s{}
if err := repo.Create(entity); err != nil {
t.Fatalf("create failed: %%v", err)
}
fetched, err := repo.GetByID(entity.ID)
if err != nil {
t.Fatalf("get failed: %%v", err)
}
if fetched.ID != entity.ID {
t.Fatalf("expected ID %%d got %%d", entity.ID, fetched.ID)
}
if err := repo.Update(entity); err != nil {
t.Fatalf("update failed: %%v", err)
}
items, err := repo.List()
if err != nil {
t.Fatalf("list failed: %%v", err)
}
if len(items) != 1 {
t.Fatalf("expected 1 item got %%d", len(items))
}
if err := repo.Delete(entity.ID); err != nil {
t.Fatalf("delete failed: %%v", err)
}
}
`, name, domainImport, capName)
}

func (g *TemplateGenerator) generateServiceTestFile(name string) string {
	capName := Capitalize(name)
	pluralName := Pluralize(capName)
	domainImport := g.moduleImportPath("internal/adapters/http/modules", name, "domain")
	return fmt.Sprintf(`// @kthulu:test:service:%[1]s
package service

import (
"testing"

"%[2]s"
)

type fake%[3]sRepository struct {
store  map[uint]*domain.%[3]s
nextID uint
}

func newFake%[3]sRepository() *fake%[3]sRepository {
return &fake%[3]sRepository{
store:  make(map[uint]*domain.%[3]s),
nextID: 1,
}
}

func (r *fake%[3]sRepository) Create(entity *domain.%[3]s) error {
if entity.ID == 0 {
entity.ID = r.nextID
r.nextID++
}
r.store[entity.ID] = entity
return nil
}

func (r *fake%[3]sRepository) GetByID(id uint) (*domain.%[3]s, error) {
return r.store[id], nil
}

func (r *fake%[3]sRepository) Update(entity *domain.%[3]s) error {
r.store[entity.ID] = entity
return nil
}

func (r *fake%[3]sRepository) Delete(id uint) error {
delete(r.store, id)
return nil
}

func (r *fake%[3]sRepository) List() ([]*domain.%[3]s, error) {
items := make([]*domain.%[3]s, 0, len(r.store))
for _, item := range r.store {
items = append(items, item)
}
return items, nil
}

func Test%[3]sServiceCRUD(t *testing.T) {
repo := newFake%[3]sRepository()
service := New%[3]sService(repo)
entity := &domain.%[3]s{}
if err := service.Create%[3]s(entity); err != nil {
t.Fatalf("create failed: %%v", err)
}
if entity.ID == 0 {
t.Fatal("expected ID to be set")
}
if _, err := service.Get%[3]sByID(entity.ID); err != nil {
t.Fatalf("get failed: %%v", err)
}
if err := service.Update%[3]s(entity); err != nil {
t.Fatalf("update failed: %%v", err)
}
items, err := service.List%[4]s()
if err != nil {
t.Fatalf("list failed: %%v", err)
}
if len(items) != 1 {
t.Fatalf("expected 1 item got %%d", len(items))
}
if err := service.Delete%[3]s(entity.ID); err != nil {
t.Fatalf("delete failed: %%v", err)
}
}
`, name, domainImport, capName, pluralName)
}

func (g *TemplateGenerator) generateHandlerTestFile(name string) string {
	capName := Capitalize(name)
	pluralName := Pluralize(capName)
	domainImport := g.moduleImportPath("internal/adapters/http/modules", name, "domain")
	return fmt.Sprintf(`// @kthulu:test:handlers:%[1]s
package handlers

import (
"encoding/json"
"fmt"
"net/http"
"net/http/httptest"
"strings"
"testing"

"github.com/gorilla/mux"
"%[2]s"
)

type fake%[3]sService struct {
store  map[uint]*domain.%[3]s
nextID uint
}

func newFake%[3]sService() *fake%[3]sService {
return &fake%[3]sService{store: make(map[uint]*domain.%[3]s), nextID: 1}
}

func (s *fake%[3]sService) Create%[3]s(entity *domain.%[3]s) error {
if entity.ID == 0 {
entity.ID = s.nextID
s.nextID++
}
s.store[entity.ID] = entity
return nil
}

func (s *fake%[3]sService) Get%[3]sByID(id uint) (*domain.%[3]s, error) {
if entity, ok := s.store[id]; ok {
return entity, nil
}
return nil, fmt.Errorf("not found")
}

func (s *fake%[3]sService) Update%[3]s(entity *domain.%[3]s) error {
s.store[entity.ID] = entity
return nil
}

func (s *fake%[3]sService) Delete%[3]s(id uint) error {
delete(s.store, id)
return nil
}

func (s *fake%[3]sService) List%[4]s() ([]*domain.%[3]s, error) {
items := make([]*domain.%[3]s, 0, len(s.store))
for _, entity := range s.store {
items = append(items, entity)
}
return items, nil
}

func Test%[3]sHandlerCRUD(t *testing.T) {
service := newFake%[3]sService()
handler := New%[3]sHandler(service)
router := mux.NewRouter()
handler.RegisterRoutes(router)

createReq := httptest.NewRequest(http.MethodPost, "/%[1]s", strings.NewReader(`+"`{}`"+`))
createRec := httptest.NewRecorder()
router.ServeHTTP(createRec, createReq)
if createRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %%d", createRec.Code)
}
var created domain.%[3]s
if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
t.Fatalf("decode failed: %%v", err)
}

listReq := httptest.NewRequest(http.MethodGet, "/%[1]s", nil)
listRec := httptest.NewRecorder()
router.ServeHTTP(listRec, listReq)
if listRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %%d", listRec.Code)
}

getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%[1]s/%%d", created.ID), nil)
getRec := httptest.NewRecorder()
router.ServeHTTP(getRec, getReq)
if getRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %%d", getRec.Code)
}

updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/%[1]s/%%d", created.ID), strings.NewReader(`+"`{}`"+`))
updateRec := httptest.NewRecorder()
router.ServeHTTP(updateRec, updateReq)
if updateRec.Code != http.StatusOK {
t.Fatalf("expected 200 got %%d", updateRec.Code)
}

deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/%[1]s/%%d", created.ID), nil)
deleteRec := httptest.NewRecorder()
router.ServeHTTP(deleteRec, deleteReq)
if deleteRec.Code != http.StatusNoContent {
t.Fatalf("expected 204 got %%d", deleteRec.Code)
}
}
`, name, domainImport, capName, pluralName)
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
