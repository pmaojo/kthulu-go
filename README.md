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

> **Quickstart tip:** When the frontend is running at `http://localhost:5173`, the header displays a connectivity badge ("API Conectada" / "Sin conexión"). Use it to confirm the UI can reach the Go API before exploring the panels.

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

## Frontend Experience

The main UI (see `frontend/src/pages/Index.tsx`) organizes functionality into dedicated panels that can be opened from the sidebar:

- **Service Canvas** — A ReactFlow-driven modeling surface for designing services, entities, actors, workflows, and their relationships.
- **Terminal** — Embedded command interface for backend and generator operations.
- **Code Editor** — In-browser editor for inspecting and adjusting generated artifacts.
- **Dashboard Preview** — High-level KPIs and activity metrics.
- **Module Catalog** — Lists available backend/service modules fetched from the API.
- **Component Scaffolder** — Launches component generation flows.
- **Template Manager** — Manages template registries, renders, and cache operations.
- **Audit Workbench** — Runs architecture and compliance audits.
- **AI Chat / Assistant** — Conversational helpers for planning, refactoring, or code suggestions.
- **Project Generator Dialog** — Accessible from the header “Generar” button for end-to-end project scaffolding.

### Service Canvas workflow

The `ServiceCanvas` component wraps the ReactFlow provider to render typed nodes (`service`, `entity`, `usecase`, `actor`, `workflow`) with a cyberpunk-styled background. A floating toolbar (`CanvasToolbar`) exposes actions to:

1. **Add nodes** with predefined payloads (e.g., new services or entities) at randomized positions.
2. **Apply templates** that hydrate the canvas with curated node/edge collections.
3. **Clear the canvas** (resets ReactFlow state and underlying store) to start fresh.
4. **Fit view** to recenter the viewport around current content.

On mount, the canvas attempts to load module definitions via `kthuluApi.listModules()`. Successful responses replace the default sample graph with server-provided service nodes. Failures fall back to the local sample graph and surface a toast explaining that the API could not be reached. Users can connect nodes by dragging handles, navigate with the minimap/controls, and inspect details in the properties side panel triggered from the header.

### Backend integrations required by the UI

The frontend’s API client (`frontend/src/services/kthuluApi.ts`) targets the Go backend at `http://localhost:8080`. Ensure the corresponding services are running so each panel functions correctly:

- **System health** — `GET /health` powers the connection badge and status checks.
- **Project planning & generation** — `POST /api/v1/projects/plan` and `POST /api/v1/projects` feed the project generator dialog.
- **Module catalog & planning** — Module listing, detail, validation, and injection routes (`/api/v1/modules/...`) back the Service Canvas and catalog views.
- **Component generation lifecycle** — CRUD endpoints under `/api/v1/components` support the scaffolder UI.
- **Template registry** — Extensive `/api/v1/templates` operations (list, render, cache, registry management, sync, verify) enable the Template Manager and toolbar templates.
- **AI services** — `/api/v1/ai/*` endpoints supply suggestions and provider management for AI Chat/Assistant features.
- **Audit engine** — `POST /api/v1/audit` runs audit scenarios surfaced in the workbench.
- **Security configuration** — `GET/PUT /api/v1/security/config` provide insights and updates for security posture tooling.

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
