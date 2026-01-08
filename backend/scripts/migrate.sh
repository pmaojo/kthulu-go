#!/bin/bash

## Migration script for Kthulu
# Usage:
#   MIGRATION_TOOL=[goose|gorm] ./scripts/migrate.sh [command]
# Examples:
#   ./scripts/migrate.sh            # run goose migrations (default)
#   MIGRATION_TOOL=gorm ./scripts/migrate.sh up
#   MIGRATION_TOOL=goose ./scripts/migrate.sh status

# Automatically detects database driver and runs migrations

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Select migration tool (default goose)
MIGRATION_TOOL=${MIGRATION_TOOL:-goose}

# Default values (Optimal: SQLite by default)
DB_DRIVER=${DB_DRIVER:-sqlite}
DATABASE_URL=${DATABASE_URL:-./kthulu.db}
MIGRATIONS_DIR=${MIGRATIONS_DIR:-./migrations}

# Extra options passed to goose
GOOSE_OPTS=${GOOSE_OPTS:-}

echo "🐙 Kthulu Migration Tool (Optimal)"
echo "Tool: $MIGRATION_TOOL"
echo "Driver: $DB_DRIVER"
echo "Database: $DATABASE_URL"
echo "Migrations: $MIGRATIONS_DIR"
echo ""

# Default command is 'up' if none provided
COMMAND=${1:-up}
shift || true

if [ "$MIGRATION_TOOL" = "gorm" ]; then
    go run ./internal/migrations/gormmigrate "$COMMAND" "$@"
    exit
fi

# Validate goose is installed
if ! command -v goose &> /dev/null; then
    echo "❌ goose is not installed. Installing..."
    go install github.com/pressly/goose/v3/cmd/goose@latest
fi

# Run migration based on driver
case $DB_DRIVER in
    sqlite)
        echo "🗃️  Running SQLite migrations..."
        goose $GOOSE_OPTS -dir "$MIGRATIONS_DIR" sqlite "$DATABASE_URL" "$COMMAND" "$@"
        ;;
    postgres)
        echo "🐘 Running PostgreSQL migrations..."
        goose $GOOSE_OPTS -dir "$MIGRATIONS_DIR" postgres "$DATABASE_URL" "$COMMAND" "$@"
        ;;
    *)
        echo "❌ Unsupported database driver: $DB_DRIVER"
        echo "Supported drivers: sqlite, postgres"
        exit 1
        ;;
esac

echo "✅ Migration completed successfully!"
