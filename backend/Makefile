# Kthulu Backend Makefile - Optimal
# SQLite by default, PostgreSQL when needed

.PHONY: help dev build test clean migrate migrate-up migrate-down migrate-status migrate-create deps

# Default target
help: ## Show this help message
	@echo "🐙 Kthulu Backend - Optimal (SQLite by default)"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
dev: deps migrate-up ## Start development server with SQLite
	@echo "🚀 Starting Kthulu in development mode (SQLite)..."
	go run cmd/service/main.go

dev-postgres: ## Start development server with PostgreSQL
	@echo "🚀 Starting Kthulu in development mode (PostgreSQL)..."
	DB_DRIVER=postgres go run cmd/service/main.go

# Build
build: ## Build the application binary
	@echo "🔨 Building Kthulu..."
	go build -o kthulu-app cmd/service/main.go

build-static: ## Build static binary (for deployment)
	@echo "🔨 Building static Kthulu binary..."
	CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o kthulu-app cmd/service/main.go

# Testing
test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-contracts: ## Run contract tests
	@echo "🔬 Running contract tests..."
	../scripts/test-contracts.sh

test-contracts-coverage: ## Run contract tests with coverage
	@echo "🔬 Running contract tests with coverage..."
	../scripts/test-coverage.sh

test-integration: ## Run integration tests
	@echo "🔧 Running integration tests..."
	../scripts/test-integration.sh

test-e2e: ## Run end-to-end tests
	@echo "🧪 Running E2E tests..."
	../scripts/test-e2e.sh

test-all: test test-contracts test-integration ## Run all tests (unit, contract, integration)
	@echo "✅ All tests completed!"

# Dependencies
deps: ## Install dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# Database migrations (Optimal - works with both SQLite and PostgreSQL)
migrate-up: ## Run database migrations up
	@echo "⬆️  Running migrations up..."
	./scripts/migrate.sh up

migrate-down: ## Run database migrations down
	@echo "⬇️  Running migrations down..."
	./scripts/migrate.sh down

migrate-status: ## Show migration status
	@echo "📊 Migration status..."
	./scripts/migrate.sh status

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_something)
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Please provide a migration name: make migrate-create NAME=create_something"; \
		exit 1; \
	fi
	@echo "📝 Creating migration: $(NAME)..."
	./scripts/migrate.sh create $(NAME) sql

migrate-reset: ## Reset database (down all + up all)
	@echo "🔄 Resetting database..."
	./scripts/migrate.sh reset

# PostgreSQL specific commands
migrate-up-postgres: ## Run migrations on PostgreSQL
	@echo "🐘 Running PostgreSQL migrations up..."
	DB_DRIVER=postgres ./scripts/migrate.sh up

migrate-down-postgres: ## Run migrations down on PostgreSQL  
	@echo "🐘 Running PostgreSQL migrations down..."
	DB_DRIVER=postgres ./scripts/migrate.sh down

# SQLite specific commands  
migrate-up-sqlite: ## Run migrations on SQLite (default)
	@echo "🗃️  Running SQLite migrations up..."
	DB_DRIVER=sqlite ./scripts/migrate.sh up

migrate-down-sqlite: ## Run migrations down on SQLite
	@echo "🗃️  Running SQLite migrations down..."
	DB_DRIVER=sqlite ./scripts/migrate.sh down

# Utilities
clean: ## Clean build artifacts and database files
	@echo "🧹 Cleaning up..."
	rm -f kthulu-app
	rm -f *.db
	rm -f coverage.out coverage.html
	rm -f server.log

clean-db: ## Clean only database files (SQLite)
	@echo "🗑️  Cleaning database files..."
	rm -f *.db

# Docker
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t kthulu:latest .

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 kthulu:latest

# Development helpers
fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	go fmt ./...

lint: ## Run linter
	@echo "🔍 Running linter..."
	golangci-lint run

# Quick start (Optimal)
setup: deps migrate-up ## Quick setup for new developers
	@echo "✅ Kthulu setup complete!"
	@echo "💡 Run 'make dev' to start development server"
	@echo "💡 Database: SQLite (./kthulu.db)"
	@echo "💡 To use PostgreSQL: 'make dev-postgres'"

# Production deployment
deploy-prep: clean build-static ## Prepare for production deployment
	@echo "📦 Production build ready: ./kthulu-app"
	@echo "💡 Copy kthulu-app binary and run migrations on target server"