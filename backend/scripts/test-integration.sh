#!/bin/bash

# @kthulu:core
# Backend integration testing script

set -e

echo "🔧 Running backend integration tests..."

# Change to backend directory
cd backend

# Set test environment variables
export NODE_ENV=test
export DB_DRIVER=sqlite
export DB_URL=":memory:"
export JWT_SECRET="test-jwt-secret-for-integration-tests"
export JWT_REFRESH_SECRET="test-jwt-refresh-secret-for-integration-tests"
export LOG_LEVEL=error

# Clean previous test results
rm -f integration-coverage.out integration-results.json

echo "📊 Running integration tests with coverage..."

# Run integration tests
go test -v -tags=integration ./internal/integration/... \
    -coverprofile=integration-coverage.out \
    -covermode=atomic \
    -json > integration-results.json 2>&1

# Generate HTML coverage report
echo "📈 Generating integration test coverage report..."
go tool cover -html=integration-coverage.out -o integration-coverage.html

# Display coverage summary
echo "📋 Integration Test Coverage Summary:"
go tool cover -func=integration-coverage.out | tail -1

# Check coverage threshold (70% minimum for integration tests)
COVERAGE=$(go tool cover -func=integration-coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
THRESHOLD=70

echo "📊 Integration Test Coverage: ${COVERAGE}%"

if (( $(echo "$COVERAGE >= $THRESHOLD" | bc -l) )); then
    echo "✅ Integration test coverage meets threshold of ${THRESHOLD}%"
    exit 0
else
    echo "❌ Integration test coverage ${COVERAGE}% is below threshold of ${THRESHOLD}%"
    exit 1
fi