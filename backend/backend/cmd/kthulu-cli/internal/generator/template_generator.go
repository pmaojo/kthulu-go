package generator

import (
	"fmt"
	"os"
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

	return nil
}

// generateMainFile generates the main.go file
func (g *TemplateGenerator) generateMainFile() string {
	return fmt.Sprintf(`// @kthulu:project:%s
// @kthulu:generated:true
// @kthulu:features:%s
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"

	"github.com/pmaojo/kthulu-go/backend/internal/core"
%s
)

func main() {
	ctx := context.Background()
	
	app := fx.New(
		// Core providers
		core.CoreRepositoryProviders(),
		
		// Module providers
%s
		
		// HTTP server
		fx.Invoke(func(lc fx.Lifecycle) {
			server := &http.Server{
				Addr:    ":8080",
				Handler: setupRoutes(),
			}
			
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					log.Println("Starting server on :8080")
					go server.ListenAndServe()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					log.Println("Stopping server")
					return server.Shutdown(ctx)
				},
			})
		}),
	)
	
	// Start application
	if err := app.Start(ctx); err != nil {
		log.Fatal("Failed to start application:", err)
	}
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// Stop application
	if err := app.Stop(ctx); err != nil {
		log.Fatal("Failed to stop application:", err)
	}
	
	log.Println("Server stopped")
}

func setupRoutes() http.Handler {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Add module routes here
%s
	
	return mux
}
`, g.config.ProjectName,
		strings.Join(g.config.Features, ","),
		g.generateModuleImports(),
		g.generateModuleProviders(),
		g.generateModuleRoutes())
}

// generateGoMod generates the go.mod file
func (g *TemplateGenerator) generateGoMod() string {
	return fmt.Sprintf(`module %s

go 1.21

require (
	go.uber.org/fx v1.20.0
	github.com/gorilla/mux v1.8.0
	gorm.io/gorm v1.25.5
	gorm.io/driver/%s v1.5.4
	github.com/golang-jwt/jwt/v5 v5.2.0
%s
)
`, g.config.ProjectName, g.config.Database, g.generateDependencies())
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

// Helper methods for code generation
func (g *TemplateGenerator) generateModuleImports() string {
	var imports []string
	// Use resolved dependencies, not just initial features
	plan, _ := g.resolver.ResolveDependencies(g.config.Features)
	for _, module := range plan.RequiredModules {
		imports = append(imports, fmt.Sprintf(`	"github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/%s"`, module))
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
		capModule := Capitalize(module)
		routes = append(routes, fmt.Sprintf(`	// %s routes`, module))
		routes = append(routes, fmt.Sprintf(`	%sHandler := %s.New%sHandler(%sService)`, module, module, capModule, module))
		routes = append(routes, fmt.Sprintf(`	%sHandler.RegisterRoutes(mux.PathPrefix("/api/v1").Subrouter())`, module))
	}
	return strings.Join(routes, "\n")
}

func (g *TemplateGenerator) generateDependencies() string {
	deps := []string{}

	if g.config.Enterprise {
		deps = append(deps,
			"	github.com/prometheus/client_golang v1.17.0",
			"	go.opentelemetry.io/otel v1.21.0",
		)
	}

	if g.config.Frontend == "react" {
		deps = append(deps, "	github.com/gorilla/websocket v1.5.0")
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
	return fmt.Sprintf(`// @kthulu:module:%s
// @kthulu:category:%s
package %s

import "go.uber.org/fx"

// Providers returns the Fx providers for the %s module
func Providers() fx.Option {
	return fx.Options(
		fx.Provide(
			New%sRepository,
			New%sService,
			New%sHandler,
		),
	)
}
`, name, info.Category, name, name,
		Capitalize(name), Capitalize(name), Capitalize(name))
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
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/%s/domain"
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
`, name, name, capName, capName, capName,
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
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/%s/domain"
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
`, name, name, capName, capName, capName,
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
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/%s/domain"
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
`, name, name, capName, capName, capName,
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
