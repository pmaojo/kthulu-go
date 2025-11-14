# Kthulu Forge — Full-Stack ERP Platform

A monolithic full-stack application combining a **Go backend** and **React frontend** with a **Bun-powered** build toolchain.

## Project Structure

```
kthulu-forge/
├── backend/              # Go backend (REST API, database, business logic)
├── frontend/             # React TypeScript frontend (Vite, Bun, Tailwind)
├── docker-compose.yml    # Local development stack (Bun, Go, PostgreSQL)
├── Dockerfile.fullstack  # Multi-stage build for production binary
└── README.md            # This file
```

## Getting Started

### Prerequisites

- **Docker** & **Docker Compose** (for full-stack development)
- **Bun** (for local frontend development) — [install](https://bun.sh)
- **Go 1.22+** (for backend development) — [install](https://golang.org)

### Quick Start (Docker)

Start the entire stack with one command:

```sh
# From repository root
docker compose up --build
```

This runs:
- PostgreSQL database on `${DB_PORT}`
- Go API server on `${API_PORT}`
- Vite dev server (via Bun) on `${WEB_PORT}` (default 5173)

See [DOCKER.md](./DOCKER.md) for more options.

### Local Development

#### Frontend (Bun + Vite + React)

```sh
cd frontend
bun install          # Install dependencies
bun run dev          # Start dev server (hot reload)
bun run build        # Build for production
bun test             # Run tests
bun run lint         # Lint code
```

#### Backend (Go)

```sh
cd backend/backend
go run ./cmd/service  # Run API server
go test ./...        # Run tests
```

#### Database

```sh
cd backend
make migrate-up       # Apply migrations
make db-ping         # Test connection
```

## Technology Stack

### Frontend
- **Vite** — Fast build tool
- **React 18** — UI library
- **TypeScript** — Type safety
- **Tailwind CSS** — Styling
- **shadcn/ui** — Component library
- **React Router** — Routing
- **Bun** — Runtime & package manager

### Backend
- **Go 1.22** — Server language
- **PostgreSQL** — Database
- **Chi** — HTTP router
- **GORM** — ORM
- **JWT** — Authentication

### DevOps
- **Docker** — Containerization
- **Docker Compose** — Local orchestration
- **Kustomize** — Kubernetes configs
- **Playwright** — E2E testing

## Development Commands

### From Project Root

```sh
# Full stack (Docker)
docker compose up --build

# Build production binary with embedded frontend
docker build -f Dockerfile.fullstack -t kthulu:latest .
```

### From Backend Directory

```sh
cd backend/backend
make dev               # Start full stack via Docker Compose
make test              # Run backend + frontend tests
make build-backend     # Build Go binary
make migrate-up        # Apply database migrations
make lint              # Lint code
```

### From Frontend Directory

```sh
cd frontend
bun install            # Install dependencies
bun run dev            # Start Vite dev server
bun run build          # Build for production
bun test               # Run Vitest tests
bun run lint           # Lint TypeScript
```

## API Documentation

Once the server is running:

```
http://localhost:${API_PORT}/docs
```

## Contributing

See [backend/CONTRIBUTING.md](./backend/CONTRIBUTING.md) for guidelines.

## License

MIT — see [backend/LICENSE](./backend/LICENSE)
