# Docker / Fullstack Quick Start

This project provides a root-level `docker-compose.yml` that will run the whole stack (DB, API, and web) with hot reload for development. All Node.js tooling now uses **Bun** for speed and efficiency.

## Development (hot reload)

Start everything from the repo root:

```bash
# From repository root
docker compose up --build
```

Services:
- **web**: Uses `oven/bun:latest` for the frontend service and mounts `./frontend` for Vite hot-reload
- **api**: Uses `golang:1.21` and mounts `./backend/backend` (Go source)
- **db**: PostgreSQL 15 for persistence

Environment variables are read from `./backend/.env` (copy `./backend/.env.example`).

- API will be accessible at: `http://localhost:${API_PORT}`
- Frontend will be accessible at: `http://localhost:${WEB_PORT}`

## Build single binary with embedded frontend

If you'd like a single binary with built frontend assets embedded, there is a root-level build Dockerfile `Dockerfile.fullstack` that will:

1. Build the frontend using Bun from `./frontend/`
2. Copy the built assets into the backend `public` folder
3. Build the Go binary

Build it using:

```bash
# Build single image
docker build -f Dockerfile.fullstack -t kthulu-fullstack:latest .

# Run the image
docker run --rm -p 8080:8080 -e HTTP_ADDR=":"${API_PORT} kthulu-fullstack:latest
```

## Notes

- Root `Makefile` targets the root-level Docker Compose. Run `make dev` to start the stack.
- All npm commands have been replaced with Bun (`bun run`, `bun install`, `bunx`).
