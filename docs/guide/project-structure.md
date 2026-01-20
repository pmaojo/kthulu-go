---
title: Project Structure
description: Understanding the Modular Monolith Architecture of Kthulu applications.
---

Kthulu generates projects following **Modular Monolith Architecture** with Vertical Slices. This ensures that your business logic is organized by feature, not technical layer.

## Directory Layout

```text
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── modules/         # Vertical Slices (Feature Modules)
│   │   ├── user/        # User module
│   │   │   ├── api/     # HTTP Handlers
│   │   │   ├── core/    # Domain & Services
│   │   │   └── store/   # Data Persistence
│   │   └── auth/        # Auth module
│   ├── views/           # GTH Frontend (Templ + HTMX)
│   │   ├── layouts/     # Base and admin layouts
│   │   ├── components/  # Reusable table, form, modal
│   │   ├── pages/       # Full page templates
│   │   └── partials/    # HTMX partial responses
│   └── infrastructure/  # Config, Logger, Middleware
└── kthulu-plan.yaml     # Project definition
```

## Modules (Vertical Slices)

Located in `internal/modules`, each module contains all layers for a single feature:

- **api/**: HTTP handlers and DTOs
- **core/**: Domain entities, interfaces, and business logic
- **store/**: Database repository implementations

## GTH Frontend

Located in `internal/views`, the frontend uses Go + Templ + HTMX:

- **layouts/**: Base HTML with HTMX CDN and CSS design system
- **components/**: Reusable table, form, and modal components
- **pages/**: Full CRUD pages for each module
- **partials/**: HTML fragments for HTMX dynamic updates

See [GTH Frontend Guide](./gth-frontend.md) for details.

## Dependency Injection

Kthulu uses [Uber fx](https://uber-go.github.io/fx/) for dependency injection. Modules are wired together via providers in each module's `module.go` file.
