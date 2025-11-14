.PHONY: dev test gen-types lint ci deploy-k8s

# Run the full development stack with hot reload
# Services defined in docker-compose.yml already watch source files
# via volume mounts, so compose up is enough for hot reload.
dev:
	docker compose up --build

# Execute backend and frontend test suites
# Uses existing bun command and go tooling
BUN = bun --cwd ./frontend
ENV ?= dev

test:
	cd backend/backend && go test ./...
	$(BUN) test

# Lint Go and TypeScript sources
lint:
	cd backend/backend && go vet ./...
	$(BUN) run lint

# Generate OpenAPI spec server stubs and TypeScript types
# Relies on oapi-codegen and associated configs
GEN_OPENAPI_CMD = go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
OPENAPI_SPEC = backend/api/openapi.yaml
GEN_GO_OUTPUT = backend/backend/internal/handlers/auto_gen.go
GEN_TS_CONFIG = frontend/oapi-ts.config.yaml

# Requires OPENAPI_SPEC and GEN_TS_CONFIG to exist
# Avoids duplication by calling the generator once for each language
# Uses npx to resolve local or global binary

gen-types:
	$(GEN_OPENAPI_CMD) --generate types,chi-server -o $(GEN_GO_OUTPUT) $(OPENAPI_SPEC)
	npx oapi-codegen --config $(GEN_TS_CONFIG) $(OPENAPI_SPEC)

# Run linting, tests and type generation
ci: lint test gen-types

# Database migration commands
migrate-up:
	cd backend && go run ./cmd/migrate -action=up

migrate-down:
	cd backend && go run ./cmd/migrate -action=down

migrate-reset:
	cd backend && go run ./cmd/migrate -action=reset

migrate-status:
	cd backend && go run ./cmd/migrate -action=status

migrate-validate:
	cd backend && go run ./cmd/migrate -action=validate

# Migrate to specific version (usage: make migrate-version VERSION=123)
migrate-version:
	cd backend && go run ./cmd/migrate -action=version -version=$(VERSION)

# Build migration tool
build-migrate:
	cd backend && go build -o bin/migrate ./cmd/migrate

# Database connection test
db-ping:
	cd backend && go run -c 'package main; import ("backend/core"; "log"); func main() { cfg, _ := core.NewConfig(); logger, _ := core.NewLogger(cfg); db, err := core.NewDB(cfg, logger); if err != nil { log.Fatal(err) }; if err := core.HealthCheck(db); err != nil { log.Fatal(err) }; log.Println("Database connection successful") }'

# Build single binary with embedded frontend
build-fullstack:
	./scripts/build-fullstack.sh

# Quick build for development (backend only)
build-backend:
	cd backend && go build -o kthulu-app ./cmd/service

# Build frontend only
build-frontend:
	cd frontend && bun run build

# Development targets - run separately in different terminals
dev-frontend: ## Start frontend dev server (Terminal 1)
	cd frontend && bun install && bun run dev

dev-backend: ## Start backend server (Terminal 2) - requires make dev-db running
	cd backend/backend && go run ./cmd/service

dev-db: ## Start PostgreSQL + Redis in Docker (Terminal 1 or background)
	docker compose up db redis

dev-setup: ## Install frontend dependencies (run once before dev-frontend)
	cd frontend && bun install

dev-local: ## Full local setup: Terminal 1: make dev-db, Terminal 2: make dev-backend, Terminal 3: make dev-frontend
	@echo "🐙 Kthulu Local Development Setup"
	@echo ""
	@echo "Run these commands in separate terminals:"
	@echo ""
	@echo "Terminal 1: make dev-db"
	@echo "Terminal 2: make dev-backend"
	@echo "Terminal 3: make dev-frontend"
	@echo ""

# Clean build artifacts
clean:
	rm -rf backend/backend/public
	rm -f backend/backend/kthulu-app
	rm -f backend/backend/deployment-info.txt
	cd frontend && rm -rf dist .bun

# Production build with optimizations
build-prod: clean
	cd frontend && bun install --production=false
	cd frontend && bun run build
	cd backend/backend && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o kthulu-app ./cmd/service

# Deploy Kubernetes resources
deploy-k8s: ## Deploy Kubernetes manifests (ENV=dev by default)
	kustomize build kustomize/overlays/$(ENV) | kubectl apply -f -

# E2E Testing
setup-e2e: ## Setup E2E testing environment
	@echo "⚙️  Setting up E2E testing..."
	@cd backend/e2e && bun install && bunx playwright install

test-e2e: ## Run end-to-end tests
	@echo "🧪 Running E2E tests..."
	@./scripts/test-e2e.sh

test-integration: ## Run backend integration tests
	@echo "🔧 Running integration tests..."
	@./scripts/test-integration.sh

test-contracts: ## Run contract tests
	@echo "🔬 Running contract tests..."
	@cd backend && make test-contracts

test-all: test test-contracts test-integration test-e2e ## Run all types of tests
	@echo "✅ All tests completed!"

# Kubernetes deployment
IMAGE_REPOSITORY ?= ghcr.io/example/kthulu
IMAGE_TAG ?= latest

deploy-k8s: ## Deploy the Kthulu application to Kubernetes via kustomize
	kustomize build kustomize/overlays/dev \
		--load-restrictor LoadRestrictionsNone \
		--enable-helm \
		--helm-set image.repository=$(IMAGE_REPOSITORY) \
		--helm-set image.tag=$(IMAGE_TAG) | kubectl apply -f -

# Help
help: ## Show this help message
	@echo "🐙 Kthulu - Full-Stack ERP Application"
	@echo ""
	@echo "Available commands:"
	@grep -E '^[[:alnum:]_-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

help-e2e: ## Show E2E testing help
	@echo "E2E Testing commands:"
	@echo "  make test-e2e           - Run all E2E tests"
	@echo "  make test-integration   - Run backend integration tests"
	@echo "  make test-contracts     - Run contract tests"
	@echo "  make setup-e2e          - Setup E2E testing environment"
	@echo ""
	@echo "E2E test options:"
	@echo "  ./scripts/test-e2e.sh --no-backend     - Skip backend tests"
	@echo "  ./scripts/test-e2e.sh --no-frontend    - Skip frontend tests"
	@echo "  ./scripts/test-e2e.sh --no-performance - Skip performance tests"
	@echo "  ./scripts/test-e2e.sh --no-build       - Skip application build"
