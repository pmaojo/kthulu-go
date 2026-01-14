---
title: Deployment
description: Deploying Kthulu applications to Cloud and Docker.
---


Kthulu applications are designed with the "One-Bin" philosophy, simplifying deployment by compiling everything into a single binary. The frontend assets are embedded directly into the Go binary.

## Building for Production

To build your application as a standalone binary:

```bash
go build -tags prod -o app cmd/server/main.go
```

This binary contains your API server and your static frontend files (served from the binary itself).

## Docker Deployment

Kthulu provides robust Docker support out of the box.

### Single Container (Recommended)

The project includes a multi-stage `Dockerfile.fullstack` that builds the frontend (using Bun) and the backend (using Go), producing a minimal scratch-based image.

To build and run:

```bash
# Build the image
docker build -f Dockerfile.fullstack -t my-app .

# Run the container
docker run -p 8080:8080 -e DOMAIN=example.com my-app
```

This image is highly optimized (typically < 20MB compressed) and includes:
- AutoTLS (Let's Encrypt) when `DOMAIN` is set.
- Embedded frontend assets.
- Timezone data and CA certificates.

### Docker Compose

For development or deployments requiring a separate database, use the provided `docker-compose.yml`:

```bash
docker-compose up -d
```

This spins up:
- `api`: The Go backend.
- `web`: The Vite dev server (for hot-reloading).
- `db`: A PostgreSQL instance.

## Cloud Deployment (CLI)

The `kthulu deploy` command automates the generation of Kubernetes manifests and Docker builds.

```bash
kthulu deploy --cloud=kubernetes --namespace=production
```

This command will:
1.  Analyze your project structure.
2.  Generate a `Dockerfile` if one is missing.
3.  Build the Docker image.
4.  Generate Kubernetes manifests (`deployment.yaml`, `service.yaml`, `hpa.yaml`) in `deployments/kubernetes/`.
5.  Apply them to your cluster using `kubectl`.

## Environment Variables

Common configuration variables:

- `PORT`: HTTP port (default: 8080).
- `DOMAIN`: Domain name for AutoTLS (e.g., `myapp.com`).
- `DATABASE_URL`: Connection string for the database (PostgreSQL/SQLite).
- `ENVIRONMENT`: `production` or `development`.
- `JWT_SECRET`: Secret for signing auth tokens.
