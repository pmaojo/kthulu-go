#!/bin/bash

# @kthulu:core
# Build script for creating a single binary with embedded frontend

set -e

echo "🐙 Kthulu Fullstack Build Script"
echo "================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKEND_DIR="backend"
FRONTEND_DIR="frontend"
PUBLIC_DIR="$BACKEND_DIR/public"
BINARY_NAME="kthulu-app"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -d "$BACKEND_DIR" ] || [ ! -d "$FRONTEND_DIR" ]; then
    print_error "This script must be run from the project root directory"
    print_error "Expected directories: $BACKEND_DIR/ and $FRONTEND_DIR/"
    exit 1
fi

# Step 1: Clean previous builds
print_status "Cleaning previous builds..."
rm -rf "$PUBLIC_DIR"
rm -f "$BACKEND_DIR/$BINARY_NAME"

# Step 2: Install frontend dependencies
print_status "Installing frontend dependencies..."
cd "$FRONTEND_DIR"

if [ ! -f "package.json" ]; then
    print_error "No package.json found in $FRONTEND_DIR"
    exit 1
fi

# Check if node_modules exists, if not install
if [ ! -d "node_modules" ]; then
    print_status "Installing npm dependencies..."
    npm install
else
    print_status "Dependencies already installed, skipping..."
fi

# Step 3: Build frontend
print_status "Building frontend application..."

# Ensure vite.config.ts has the correct output directory
if [ -f "vite.config.ts" ]; then
    # Check if the config already has the correct outDir
    if ! grep -q "outDir.*public" vite.config.ts; then
        print_warning "Updating vite.config.ts to output to backend/public"
        
        # Create a backup
        cp vite.config.ts vite.config.ts.backup
        
        # Update the config (this is a simple approach, might need adjustment)
        cat > vite.config.ts << 'EOF'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  base: '/',
  build: {
    outDir: '../backend/public',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
EOF
        print_success "Updated vite.config.ts"
    fi
fi

# Build the frontend
npm run build

if [ $? -ne 0 ]; then
    print_error "Frontend build failed"
    exit 1
fi

print_success "Frontend built successfully"

# Step 4: Verify frontend build
cd ..
if [ ! -f "$PUBLIC_DIR/index.html" ]; then
    print_error "Frontend build failed - no index.html found in $PUBLIC_DIR"
    exit 1
fi

print_status "Frontend assets:"
ls -la "$PUBLIC_DIR"

# Step 5: Build backend
print_status "Building backend application..."
cd "$BACKEND_DIR"

# Ensure go.mod exists
if [ ! -f "go.mod" ]; then
    print_error "No go.mod found in $BACKEND_DIR"
    exit 1
fi

# Build the Go binary
print_status "Compiling Go binary..."
go build -o "$BINARY_NAME" ./cmd/service

if [ $? -ne 0 ]; then
    print_error "Backend build failed"
    exit 1
fi

print_success "Backend built successfully"

# Step 6: Verify the binary
if [ ! -f "$BINARY_NAME" ]; then
    print_error "Binary not found: $BINARY_NAME"
    exit 1
fi

# Get binary size
BINARY_SIZE=$(du -h "$BINARY_NAME" | cut -f1)
print_success "Binary created: $BINARY_NAME ($BINARY_SIZE)"

# Step 7: Create deployment info
print_status "Creating deployment info..."
cat > deployment-info.txt << EOF
Kthulu Fullstack Deployment
===========================

Build Date: $(date)
Binary: $BINARY_NAME
Binary Size: $BINARY_SIZE
Frontend Assets: $(ls -1 public | wc -l) files

Deployment Instructions:
1. Copy the binary: $BINARY_NAME
2. Copy the public/ directory
3. Set environment variables (see .env.example)
4. Run: ./$BINARY_NAME

The application will serve:
- API endpoints at /api/*
- Frontend application at /*
- Health check at /health
- API documentation at /docs

Default port: 8080 (configurable via HTTP_ADDR)
EOF

print_success "Deployment info created: deployment-info.txt"

# Step 8: Final summary
echo ""
echo "🎉 Build Complete!"
echo "=================="
print_success "Single binary created: $BACKEND_DIR/$BINARY_NAME"
print_success "Frontend assets embedded in: $PUBLIC_DIR/"
print_success "Ready for deployment!"
echo ""
echo "To run locally:"
echo "  cd $BACKEND_DIR"
echo "  ./$BINARY_NAME"
echo ""
echo "To deploy:"
echo "  1. Copy $BACKEND_DIR/$BINARY_NAME to your server"
echo "  2. Copy $PUBLIC_DIR/ to the same directory as the binary"
echo "  3. Set environment variables"
echo "  4. Run the binary"
echo ""
print_status "Happy coding! 🚀"