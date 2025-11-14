#!/bin/bash

# Development setup script for Kthulu
# This script sets up a local development environment

set -e

echo "🐙 Kthulu Development Setup"
echo "=========================="

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Load environment variables from .env.local or Vault
ENV_FILE=".env.local"
if [ -f "$ENV_FILE" ]; then
    echo -e "${BLUE}🔐 Loading secrets from $ENV_FILE${NC}"
    set -a
    # shellcheck source=/dev/null
    source "$ENV_FILE"
    set +a
elif command -v vault >/dev/null 2>&1 && [ -n "$VAULT_SECRET_PATH" ]; then
    echo -e "${BLUE}🔐 Fetching secrets from Vault path ${VAULT_SECRET_PATH}${NC}"
    if ! command -v jq >/dev/null 2>&1; then
        echo -e "${RED}❌ jq is required to parse Vault secrets${NC}"
        exit 1
    fi
    set -a
    vault kv get -format=json "$VAULT_SECRET_PATH" \
        | jq -r '.data.data | to_entries[] | "\(.key)=\(.value)"' > /tmp/.env.vault
    # shellcheck source=/dev/null
    source /tmp/.env.vault
    rm /tmp/.env.vault
    set +a
else
    echo -e "${YELLOW}⚠️  No .env.local file or Vault configuration found. Using defaults.${NC}"
fi

# Ensure required secrets are present
if [ -z "$JWT_SECRET" ] || [ -z "$JWT_REFRESH_SECRET" ]; then
    echo -e "${RED}❌ JWT_SECRET and JWT_REFRESH_SECRET must be set${NC}"
    exit 1
fi

# Default database credentials if not provided
POSTGRES_DB=${POSTGRES_DB:-kthulu}
POSTGRES_USER=${POSTGRES_USER:-kthulu}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-kthulu}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

echo -e "${BLUE}🐳 Setting up PostgreSQL database...${NC}"

# Stop and remove existing container if it exists
docker stop kthulu-postgres 2>/dev/null || true
docker rm kthulu-postgres 2>/dev/null || true

# Start PostgreSQL container
docker run -d \
    --name kthulu-postgres \
    -e POSTGRES_DB="$POSTGRES_DB" \
    -e POSTGRES_USER="$POSTGRES_USER" \
    -e POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
    -p 5432:5432 \
    postgres:15-alpine

echo -e "${YELLOW}⏳ Waiting for PostgreSQL to be ready...${NC}"
sleep 5

# Test database connection
max_attempts=30
attempt=1
while [ $attempt -le $max_attempts ]; do
    if docker exec kthulu-postgres pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ PostgreSQL is ready!${NC}"
        break
    fi
    
    if [ $attempt -eq $max_attempts ]; then
        echo -e "${RED}❌ PostgreSQL failed to start after $max_attempts attempts${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}⏳ Attempt $attempt/$max_attempts - waiting for PostgreSQL...${NC}"
    sleep 2
    ((attempt++))
done

echo -e "${BLUE}📦 Installing backend dependencies...${NC}"
cd backend
go mod download

echo -e "${BLUE}🗄️ Running database migrations...${NC}"
go run ./cmd/migrate up

echo -e "${BLUE}📦 Installing frontend dependencies...${NC}"
cd ../frontend
npm install

cd ..

echo -e "${BLUE}🚀 Starting development servers with hot reload...${NC}"
trap 'kill $(jobs -p)' EXIT

(
    cd backend
    air
) &

(
    cd frontend
    npm run dev
) &

wait

echo -e "${GREEN}🎉 Development environment setup complete!${NC}"
echo ""
echo -e "${BLUE}To start the development servers:${NC}"
echo ""
echo -e "${YELLOW}Backend (API):${NC}"
echo "  cd backend"
echo "  go run ./cmd/service"
echo ""
echo -e "${YELLOW}Frontend (React):${NC}"
echo "  cd frontend"
echo "  npm run dev"
echo ""
echo -e "${YELLOW}Or build the fullstack binary:${NC}"
echo "  make build-fullstack"
echo "  cd backend"
echo "  ./kthulu-app"
echo ""
echo -e "${BLUE}Database connection:${NC}"
echo "  Host: localhost:5432"
echo "  Database: $POSTGRES_DB"
echo "  Username: $POSTGRES_USER"
echo "  Password: $POSTGRES_PASSWORD"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"
