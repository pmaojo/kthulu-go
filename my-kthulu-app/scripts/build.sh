#!/bin/bash
# Build script for my-kthulu-app

set -e

echo "Building my-kthulu-app..."

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
