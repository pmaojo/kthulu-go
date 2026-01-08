#!/bin/bash

# Script para corregir automáticamente issues de linting comunes
# Compatible con macOS y Linux

set -e

echo "🔧 Fixing common linting issues..."

# Función para mostrar progreso
show_progress() {
    echo "  ✓ $1"
}

# 1. Formatear código con gofmt
echo "📝 Formatting code with gofmt..."
find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | xargs gofmt -w
show_progress "Code formatted"

# 2. Organizar imports con goimports
echo "📦 Organizing imports with goimports..."
if command -v goimports >/dev/null 2>&1; then
    find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | xargs goimports -w -local backend
    show_progress "Imports organized"
else
    echo "⚠️  goimports not found, installing..."
    go install golang.org/x/tools/cmd/goimports@latest
    find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | xargs goimports -w -local backend
    show_progress "Imports organized"
fi

# 3. Corregir misspellings comunes
echo "📝 Fixing common misspellings..."
find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -exec sed -i.bak 's/cancelled/canceled/g' {} \;
find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -exec sed -i.bak 's/contactos/contacts/g' {} \;
# Limpiar archivos backup
find . -name "*.go.bak" -delete
show_progress "Misspellings fixed"

# 4. Añadir constantes para strings repetidos
echo "📝 Creating constants for repeated strings..."

# Crear archivo de constantes comunes si no existe
if [ ! -f "internal/domain/common/constants.go" ]; then
    mkdir -p internal/domain/common
    cat > internal/domain/common/constants.go << 'EOF'
// @kthulu:core
package common

// Database drivers
const (
	SQLiteDriver   = "sqlite"
	PostgresDriver = "postgres"
	SQLite3Dialect = "sqlite3"
)

// Sort orders
const (
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

// Default sort fields
const (
	DefaultSortField = "created_at"
)

// Health check paths
const (
	HealthzPath = "/healthz"
)

// Status values
const (
	StatusHealthy = "healthy"
)

// Security levels
const (
	SecurityLevelLow    = "LOW"
	SecurityLevelMedium = "MEDIUM"
	SecurityLevelHigh   = "HIGH"
)
EOF
    show_progress "Constants file created"
fi

# 5. Ejecutar go mod tidy para limpiar dependencias
echo "📦 Cleaning up dependencies..."
go mod tidy
show_progress "Dependencies cleaned"

echo ""
echo "✅ Linting fixes completed!"
echo ""
echo "📊 Running linter to check remaining issues..."
echo "   (Note: Some issues require manual intervention)"
echo ""

# Ejecutar linter con límite de issues para no abrumar
golangci-lint run --max-issues-per-linter=10 --max-same-issues=3 || true

echo ""
echo "🎯 Next steps:"
echo "   1. Review remaining linting issues above"
echo "   2. Fix complex issues manually (cognitive complexity, error handling)"
echo "   3. Run 'make lint' to verify all fixes"
echo "   4. Consider refactoring functions with high complexity"