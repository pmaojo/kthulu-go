#!/bin/bash

# @kthulu:core
# Comprehensive contract testing script

set -e

echo "🔬 Running comprehensive contract tests..."

# Change to backend directory
cd backend

# Run OAuth SSO module tests
echo "🔐 Testing OAuth SSO module..."
go test -v ./internal/modules/oauthsso/...

# Run repository contract tests
echo "📦 Testing repository contracts..."
go test -v -run TestRepositoryContracts ./internal/contracts/

echo "👤 Testing user repository contract..."
go test -v -run TestUserRepositoryContract ./internal/contracts/

echo "🏢 Testing organization repository contract..."
go test -v -run TestOrganizationRepositoryContract ./internal/contracts/

echo "📞 Testing contact repository contract..."
go test -v -run TestContactRepositoryContract ./internal/contracts/

echo "📦 Testing product repository contract..."
go test -v -run TestProductRepositoryContract ./internal/contracts/

echo "📄 Testing invoice repository contract..."
go test -v -run TestInvoiceRepositoryContract ./internal/contracts/

echo "📊 Testing inventory repository contract..."
go test -v -run TestInventoryRepositoryContract ./internal/contracts/

echo "📅 Testing calendar repository contract..."
go test -v -run TestCalendarRepositoryContract ./internal/contracts/

echo "🔑 Testing role repository contract..."
go test -v -run TestRoleRepositoryContract ./internal/contracts/

echo "🎫 Testing refresh token repository contract..."
go test -v -run TestRefreshTokenRepositoryContract ./internal/contracts/

# Run HTTP contract tests
echo "🌐 Testing HTTP endpoint contracts..."
go test -v -run TestHTTPContracts ./internal/contracts/

echo "❤️ Testing health endpoint contracts..."
go test -v -run testHealthEndpointContracts ./internal/contracts/

echo "🔐 Testing auth endpoint contracts..."
go test -v -run testAuthEndpointContracts ./internal/contracts/

echo "👤 Testing user endpoint contracts..."
go test -v -run testUserEndpointContracts ./internal/contracts/

echo "🏢 Testing organization endpoint contracts..."
go test -v -run testOrganizationEndpointContracts ./internal/contracts/

echo "📞 Testing contact endpoint contracts..."
go test -v -run testContactEndpointContracts ./internal/contracts/

echo "📦 Testing product endpoint contracts..."
go test -v -run testProductEndpointContracts ./internal/contracts/

echo "📄 Testing invoice endpoint contracts..."
go test -v -run testInvoiceEndpointContracts ./internal/contracts/

echo "📊 Testing inventory endpoint contracts..."
go test -v -run testInventoryEndpointContracts ./internal/contracts/

echo "📅 Testing calendar endpoint contracts..."
go test -v -run testCalendarEndpointContracts ./internal/contracts/

# Run response format tests
echo "📋 Testing response formats..."
go test -v -run TestEndpointResponseFormats ./internal/contracts/

echo "🔢 Testing HTTP status codes..."
go test -v -run TestHTTPStatusCodes ./internal/contracts/

echo "📄 Testing content type headers..."
go test -v -run TestContentTypeHeaders ./internal/contracts/

echo "✅ All contract tests completed successfully!"