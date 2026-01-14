---
title: Deployment
description: Deploying Kthulu applications to Cloud and Edge.
---


Kthulu applications are compiled into a single binary ("One-Bin" philosophy), which makes deployment incredibly simple. The frontend assets are embedded into the Go binary.

## Building for Production

To build your application:

```bash
go build -tags prod -o app cmd/server/main.go
```

This binary contains your API server and your static frontend files.

## Cloud Deployment (Docker)

Kthulu generates a `Dockerfile` optimized for production.

```bash
docker build -t my-app .
docker run -p 8080:8080 -e DOMAIN=example.com my-app
```

The application automatically handles AutoTLS (Let's Encrypt) when the `DOMAIN` environment variable is set.

## Wasmer Edge

Kthulu has first-class support for WebAssembly (WASI).

1.  **Compile to WASM**:
    ```bash
    GOOS=wasip1 GOARCH=wasm go build -o app.wasm cmd/server/main.go
    ```

2.  **Deploy**:
    ```bash
    wasmer deploy
    ```

The `kthulu deploy` command automates this process, generating the necessary `wasmer.toml` configuration.

## Environment Variables

Common configuration variables:

- `PORT`: HTTP port (default: 8080)
- `DOMAIN`: Domain name for AutoTLS
- `DATABASE_URL`: Connection string for the database
- `ENVIRONMENT`: `production` or `development`
