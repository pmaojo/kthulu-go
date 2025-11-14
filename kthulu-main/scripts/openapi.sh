#!/bin/bash

# @kthulu:core
# OpenAPI Specification Generation Script
# Generates OpenAPI 3.1 specification from Go handlers

set -e

BACKEND_DIR="backend"
API_DIR="api"
OPENAPI_FILE="$API_DIR/openapi.yaml"

if [ ! -d "$BACKEND_DIR" ]; then
  echo "backend directory not found" >&2
  exit 1
fi

mkdir -p "$API_DIR"

# Generate Swagger documentation from Go source
if ! command -v swag >/dev/null 2>&1; then
  go install github.com/swaggo/swag/cmd/swag@latest
fi

pushd "$BACKEND_DIR" >/dev/null
swag init -g cmd/service/main.go -o docs --parseDependency --parseInternal

# Convert Swagger 2.0 spec to OpenAPI 3.1
npx -y swagger2openapi --targetVersion=3.1.0 docs/swagger.yaml -o "../$OPENAPI_FILE"

popd >/dev/null

echo "OpenAPI specification generated at $OPENAPI_FILE"
