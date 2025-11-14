#!/bin/bash

# @kthulu:core
# Test coverage script for comprehensive contract testing

set -e

echo "🧪 Running comprehensive contract tests with coverage..."

# Change to backend directory
cd backend

# Clean previous coverage files
rm -f coverage.out coverage.html

# Run all tests with coverage
echo "📊 Running tests with coverage..."
go test -v -race -coverprofile=coverage.out -covermode=atomic ./internal/contracts/...

# Generate HTML coverage report
echo "📈 Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html

# Display coverage summary
echo "📋 Coverage Summary:"
go tool cover -func=coverage.out | tail -1

# Check coverage threshold (80% minimum)
COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
THRESHOLD=80

echo "📊 Total Coverage: ${COVERAGE}%"

if (( $(echo "$COVERAGE >= $THRESHOLD" | bc -l) )); then
    echo "✅ Coverage meets threshold of ${THRESHOLD}%"
    exit 0
else
    echo "❌ Coverage ${COVERAGE}% is below threshold of ${THRESHOLD}%"
    exit 1
fi