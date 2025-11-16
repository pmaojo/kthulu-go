# test-project Makefile

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
