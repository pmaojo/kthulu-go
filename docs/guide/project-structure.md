---
title: Project Structure
description: Understanding the Hexagonal Architecture of Kthulu applications.
---


Kthulu generates projects following **Hexagonal Architecture** (also known as Ports and Adapters). This ensures that your business logic is isolated from external concerns like databases or HTTP frameworks.

## Directory Layout

```text
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── core/            # Domain logic (Ports)
│   │   ├── user/        # User module
│   │   └── auth/        # Auth module
│   ├── adapters/        # Implementation details (Adapters)
│   │   ├── http/        # REST/gRPC handlers
│   │   └── postgres/    # Database repositories
│   └── infrastructure/  # Config, Logger, Metrics
├── frontend/            # React/Vite application
└── kthulu-plan.yaml     # Project definition
```

## Core (The Domain)

Located in `internal/core`, this is where your business rules live. It depends on nothing but the Go standard library.

- **Entities**: Pure Go structs (e.g., `User`, `Order`).
- **Ports**: Interfaces defining what the core needs (e.g., `UserRepository`, `EmailService`).
- **Use Cases**: Application logic (e.g., `CreateUser`, `ProcessOrder`).

## Adapters (The Infrastructure)

Located in `internal/adapters`, these packages implement the interfaces defined in the Core.

- **Primary Adapters**: Drive the application (e.g., HTTP Handlers, CLI commands).
- **Secondary Adapters**: Driven by the application (e.g., Postgres implementation of `UserRepository`).

## Dependency Injection

Kthulu uses [Uber fx](https://uber-go.github.io/fx/) for dependency injection. Modules are wired together in `internal/infrastructure/container/container.go`.

## Frontend

The `frontend/` directory is a standard Vite application, but organized by feature modules to match the backend structure.

- `frontend/src/modules/auth` matches `internal/core/auth`.
