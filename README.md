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

## CLI Usage

The repository ships with a comprehensive command-line interface in `backend/backend/cmd/kthulu-cli`. You can run it directly while developing or install it as a standalone binary:

```sh
# Run without installing
cd backend/backend
go run ./cmd/kthulu-cli --help

# Build and expose a reusable binary
go build -o bin/kthulu ./cmd/kthulu-cli
export PATH="$(pwd)/bin:$PATH"   # or copy the binary somewhere on your PATH
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
