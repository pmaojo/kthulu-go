package generator

import (
	"backend/cmd/kthulu-cli/internal/resolver"
	"fmt"
)

// Generate clean template functions
func (g *TemplateGenerator) generateDomainFileFixed(name string, info *resolver.ModuleInfo) string {
	capName := Capitalize(name)
	pluralName := Pluralize(capName)

	template := `// @kthulu:domain:%s
package domain

import "time"

// %s represents a %s entity
type %s struct {
	ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
	
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
`
	return fmt.Sprintf(template,
		name, capName, name, capName,
		capName, capName, capName, capName, capName, capName,
		capName, capName, capName, capName, capName, capName,
		capName, capName, pluralName, capName)
}

func (g *TemplateGenerator) generateRepositoryFileFixed(name string, info *resolver.ModuleInfo) string {
	capName := Capitalize(name)

	template := `// @kthulu:repository:%s
package repository

import (
	"gorm.io/gorm"
	"backend/internal/modules/%s/domain"
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
`
	return fmt.Sprintf(template,
		name, name, capName, capName, capName, capName,
		capName, capName, capName, capName, capName,
		capName, capName, capName, capName, capName, capName)
}

func (g *TemplateGenerator) generateServiceFileFixed(name string, info *resolver.ModuleInfo) string {
	capName := Capitalize(name)
	pluralName := Pluralize(capName)

	template := `// @kthulu:service:%s
package service

import (
	"backend/internal/modules/%s/domain"
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
`
	return fmt.Sprintf(template,
		name, name, capName, capName, capName, capName, capName, capName,
		capName, capName, capName, capName, capName, capName,
		capName, capName, capName, capName, capName,
		capName, pluralName, capName)
}

// Configuration generation functions
func (g *TemplateGenerator) generateDockerCompose() string {
	dbService := g.getDatabaseService()

	return fmt.Sprintf(`version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=%s
      - DB_PORT=%s
      - DB_NAME=%s
      - DB_USER=admin
      - DB_PASSWORD=password
    depends_on:
      - %s
    networks:
      - %s-network

  %s:
    image: %s
    environment:
      %s
    ports:
      - "%s:%s"
    networks:
      - %s-network
    volumes:
      - %s-data:/var/lib/%s

volumes:
  %s-data:

networks:
  %s-network:
    driver: bridge
`, dbService.host, dbService.port, g.config.ProjectName,
		dbService.name, g.config.ProjectName, dbService.name, dbService.image,
		dbService.env, dbService.port, dbService.port, g.config.ProjectName,
		g.config.ProjectName, dbService.name, g.config.ProjectName, g.config.ProjectName)
}

func (g *TemplateGenerator) generateMakefile() string {
	return fmt.Sprintf(`# %s Makefile

.PHONY: build test run clean docker-up docker-down migrate

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run tests
test:
	go test ./...

# Run the application
run:
	go run cmd/server/main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Start development environment
docker-up:
	docker-compose up -d

# Stop development environment  
docker-down:
	docker-compose down

# Run database migrations
migrate:
	go run cmd/migrate/main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Development setup
dev: deps docker-up migrate
	@echo "Development environment ready!"
`, g.config.ProjectName)
}

func (g *TemplateGenerator) generateAppConfig() string {
	return fmt.Sprintf(`# %s Configuration

app:
  name: %s
  version: "1.0.0"
  port: 8080
  env: development

database:
  driver: %s
  host: localhost
  port: %s
  name: %s
  user: admin
  password: password
  max_connections: 100

auth:
  jwt_secret: "your-jwt-secret-key"
  jwt_expiry: "24h"

logging:
  level: info
  format: json

%s`, g.config.ProjectName, g.config.ProjectName,
		g.config.Database, g.getDefaultPort(), g.config.ProjectName, g.getEnterpriseConfig())
}

func (g *TemplateGenerator) generateDockerfile() string {
	return `FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./main"]
`
}

func (g *TemplateGenerator) generateBuildScript() string {
	return fmt.Sprintf(`#!/bin/bash
# Build script for %s

set -e

echo "Building %s..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf bin/

# Create bin directory
mkdir -p bin/

# Build server
echo "Building server..."
go build -o bin/server cmd/server/main.go

# Build CLI tools
echo "Building CLI tools..."
go build -o bin/migrate cmd/migrate/main.go

echo "Build complete!"
echo "Server binary: bin/server"
echo "Migration tool: bin/migrate"
`, g.config.ProjectName, g.config.ProjectName)
}

// Helper functions
type DatabaseService struct {
	name  string
	image string
	env   string
	host  string
	port  string
}

func (g *TemplateGenerator) getDatabaseService() DatabaseService {
	switch g.config.Database {
	case "postgres":
		return DatabaseService{
			name:  "postgres",
			image: "postgres:15-alpine",
			env:   "POSTGRES_DB=" + g.config.ProjectName + "\n      POSTGRES_USER=admin\n      POSTGRES_PASSWORD=password",
			host:  "postgres",
			port:  "5432",
		}
	case "mysql":
		return DatabaseService{
			name:  "mysql",
			image: "mysql:8.0",
			env:   "MYSQL_DATABASE=" + g.config.ProjectName + "\n      MYSQL_USER=admin\n      MYSQL_PASSWORD=password\n      MYSQL_ROOT_PASSWORD=rootpassword",
			host:  "mysql",
			port:  "3306",
		}
	default: // sqlite
		return DatabaseService{
			name:  "app",
			image: "",
			env:   "",
			host:  "localhost",
			port:  "0",
		}
	}
}

func (g *TemplateGenerator) getDefaultPort() string {
	switch g.config.Database {
	case "postgres":
		return "5432"
	case "mysql":
		return "3306"
	default:
		return "0"
	}
}

func (g *TemplateGenerator) getEnterpriseConfig() string {
	if !g.config.Enterprise {
		return ""
	}

	return `
enterprise:
  audit:
    enabled: true
    retention_days: 365
  
  security:
    rate_limiting: true
    cors_enabled: true
    allowed_origins: ["*"]
  
  monitoring:
    enabled: true
    metrics_port: 9090
    health_check_path: "/health"
`
}
