# Kthulu Forge — AI-assisted Software Foundry

Kthulu Forge packages an AI-ready **Go backend** with a **React/Bun frontend** to help teams plan, model, and generate modular software projects. The platform exposes two primary interfaces:

- **Web UI** — a service canvas, scaffolding dashboards, and live tooling for orchestrating project blueprints.
- **CLI** — a template-driven generator and operations companion for managing services, modules, and automation pipelines.

## Who Is Kthulu Forge For?

Kthulu Forge serves multidisciplinary software delivery teams that need reliable automation without sacrificing architectural rigor:

- **Platform engineering groups** who maintain internal developer platforms and want reusable service templates, event-driven workflows, and enforceable standards.
- **Solution architects and tech leads** responsible for aligning new initiatives with reference architectures, governance requirements, and traceable decisions.
- **AI-assisted delivery teams** experimenting with generative workflows that must stay grounded in typed contracts, hexagonal boundaries, and SOLID-aligned modules.
- **Consultancies and agencies** packaging industry-specific accelerators that demand repeatable scaffolds, seeded datasets, and rapid iteration loops.

## Landing Page Messaging Guide

When designing a marketing landing page, highlight the outcomes and credibility signals that matter most to the audiences above:

1. **Hero statement** — Position Kthulu Forge as an AI-assisted foundry that fuses Web UI orchestration with CLI automation to deliver production-ready service blueprints.
2. **Core value pillars** — Emphasize generative modeling, policy-aware scaffolding, and real-time collaboration across the UI and terminal.
3. **Architecture assurances** — Call out hexagonal boundaries, SOLID module design, contract-first APIs, and quality gates (tests, linting, migrations) wired into every template.
4. **Interface spotlight** — Showcase screenshots or animations of the service canvas, terminal automation, and module catalog working together end-to-end.
5. **Proof & trust** — Include case studies, integration badges (GitHub, Kubernetes, OpenAI), and testimonials from platform or architecture leads.
6. **Call-to-action** — Offer a guided Web UI tour, CLI quickstart script, and contact form for enterprise enablement.

Teams can move from whiteboard to working service skeletons while preserving governance.

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

#### CLI installation

You can install the `kthulu-cli` binary in two ways:

1. **Descargar el artefacto publicado**. Cada tag `v*` genera una GitHub Release con binarios y checksums para Linux, macOS y Windows. Descargue el archivo correspondiente a su plataforma, verifique el checksum y añada el ejecutable a su `PATH`.
2. **`go install` (cuando el módulo tenga ruta canónica).** Ejecute `go install github.com/<org>/kthulu-go/backend/cmd/kthulu-cli@latest` para compilar y situar la herramienta en su `$GOBIN`.

Ambos métodos respetan la inyección de metadatos de versión y build realizados durante el proceso de release.

#### Database

Run these helpers from the repository root—the commands are provided by the
top-level `Makefile`, so you won't find a separate one under `backend/`.

```sh
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

## CLI Usage

The repository ships with a comprehensive command-line interface in `backend/backend/cmd/kthulu-cli`. You can run it directly while developing or install it as a standalone binary:

```sh
# Run without installing
cd backend/backend
go run ./cmd/kthulu-cli --help

# Build and expose a reusable binary
go build -o bin/kthulu ./cmd/kthulu-cli
export PATH="$(pwd)/bin:$PATH"   # or copy the binary somewhere on your PATH

# Install globally from the module path
go install github.com/kthulu/kthulu-go/backend/cmd/kthulu-cli@latest
```

### Core commands

| Command | Purpose | Example |
| --- | --- | --- |
| `kthulu create <name>` | Scaffolds an intelligent project from curated templates with optional feature toggles, database/front-end choices, and enterprise add-ons. | `kthulu create my-app --template=saas --features=user,invoice --enterprise` |
| `kthulu add module <name>` / `kthulu add component <type> <name>` | Adds new modules or components to an existing project, resolving dependencies, integrations, and optional tests/migrations. | `kthulu add module payment --with=stripe`<br>`kthulu add component handler User --with-tests` |
| `kthulu generate <type> <name>` | Generates production-ready code artifacts (handlers, use cases, entities, migrations, tests, etc.) with security, validation, and metrics toggles. | `kthulu generate handler Order --crud --auth` |
| `kthulu ai "<prompt>"` | Invokes the AI assistant to propose or apply code updates. Subcommands like `kthulu ai review` and `kthulu ai optimize` offer code review and performance tuning workflows. | `kthulu ai "Add rate limiting to the API" --provider=openai --model=gpt-4` |
| `kthulu audit` / `kthulu deploy` / `kthulu status` / `kthulu upgrade` | Enterprise tooling for auditing security & compliance, cloud deployment orchestration, project health checks, and framework upgrades. | `kthulu deploy --cloud=gcp --region=us-central1` |
| `kthulu secure` | Scans dependencies for vulnerabilities and optionally patches them (auto-committing on CI when enabled). | `kthulu secure --patch` |
| `kthulu migrate <subcommand>` | Manages database migrations (`up`, `down`, `reset`, `status`, `version`, `validate`) using the shared backend configuration. | `kthulu migrate up` |

### Templates & advanced guidance

- Project blueprints, feature snippets, and other scaffolding assets live under [`backend/backend/cmd/kthulu-cli/templates`](backend/backend/cmd/kthulu-cli/templates).
- AI workflows rely on the Gemini integration by default—set `GEMINI_API_KEY` (or use `--mock` for offline exploration) and run CLI commands from the root of a generated project so dependency discovery works correctly.
- For deeper dives into scripted workflows, Makefile integrations, and advanced scenarios, see [`backend/docs/CLI.md`](backend/docs/CLI.md) and the supplemental notes in [`backend/docs/cli/`](backend/docs/cli/).

## API Documentation

Once the server is running:

```
http://localhost:${API_PORT}/docs
```

## Contributing

See [backend/CONTRIBUTING.md](./backend/CONTRIBUTING.md) for guidelines.

## License

MIT — see [backend/LICENSE](./backend/LICENSE)
