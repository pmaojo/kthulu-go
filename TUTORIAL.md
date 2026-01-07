# Building a Car Workshop API with Kthulu

This tutorial demonstrates how to build a production-ready Car Workshop Management API using the Kthulu CLI. We will create a new project, add modules for customers, cars, services, and bookings, and address some initial frictions to get the application running.

## Prerequisites

- Go 1.22+
- Docker (optional, for running the generated Dockerfile)
- Kthulu CLI installed

### Installing Kthulu CLI

If you haven't installed the CLI yet, you can build it from source:

```bash
cd backend/backend
go build -o ../../bin/kthulu-cli ./cmd/kthulu-cli
```

*Note: Adjust the paths based on your current directory structure.*

## Step 1: Initialize the Project

We will start by creating a new project named `workshop-api`. We'll use the `monolith` template, `postgres` as the database, and include the `users` feature (for future authentication).

```bash
kthulu-cli create workshop-api \
  --template=monolith \
  --database=postgres \
  --features=users \
  --module-path=github.com/example/workshop-api
```

This command generates a complete project structure with Clean Architecture, dependency injection (Uber Fx), and configured tooling (Makefile, Dockerfile, etc.).

## Step 2: Add Modules

Now we will add the core entities for our Car Workshop application using the `add module` command.

### 2.1 Customers Module

```bash
cd workshop-api
kthulu-cli add module customers name:string email:string phone:string address:string -y
```

### 2.2 Cars Module

```bash
kthulu-cli add module cars \
  make:string \
  model:string \
  year:int \
  plate:string \
  vin:string \
  customer_id:int \
  -y
```

### 2.3 Services Module

```bash
kthulu-cli add module services \
  name:string \
  description:string \
  price:float \
  duration:int \
  -y
```

### 2.4 Bookings Module

```bash
kthulu-cli add module bookings \
  car_id:int \
  service_id:int \
  booking_date:time \
  status:string \
  -y
```

## Step 3: Fix Generated Code

During this process, we encountered a few bugs in the generated code that prevent immediate compilation and startup. We need to fix them manually:

1.  **Fix Missing Import:** Open `internal/core/providers.go` and add `"gorm.io/driver/sqlite"` to the imports. This is required for the test mode configuration generated in that file.
2.  **Fix Dependency Injection:** Open `cmd/server/main.go`. The `setupRoutes` function creates the router, but it is not provided to the Fx container.
    -   Change `users.Providers(), ...` block to explicitly provide the router:
        ```go
        // Core providers
        fx.Provide(setupRoutes),

        // Module providers
        users.Providers(), ...
        ```
    -   Update the `fx.Invoke` signature to accept `router *mux.Router` so it can be used to register routes.
3.  **Missing Files:**
    -   Create `cmd/migrate/main.go` (referenced by `make migrate`) if it's missing.
    -   Create a migration for the `users` table if not automatically generated.

## Generated Code Overview

The CLI has generated a fully functional backend structure:

- **Domain (`internal/adapters/http/modules/bookings/domain/bookings.go`)**: Defines the `Bookings` struct and interfaces.
- **Repository (`.../repository/bookings_repository.go`)**: Implements database access using GORM.
- **Service (`.../service/bookings_service.go`)**: Contains business logic.
- **Handler (`.../handlers/bookings_handler.go`)**: HTTP handlers for REST API endpoints.

## Running the Application

You can now run the application:

```bash
go mod tidy
go run cmd/server/main.go
```

### API Endpoints

Due to current generation logic, there is a mix of routing paths:

-   **Users:** `GET /api/v1/users` (Configured manually in `main.go`)
-   **Bookings:** `GET /bookingss` (Auto-registered at root)
-   **Customers:** `GET /customerss`
-   **Cars:** `GET /carss`
-   **Services:** `GET /servicess`

---

## Roadmap: Fixing Frictions

To streamline the developer experience, the following improvements are planned for the Kthulu CLI:

### 1. Fix Broken Code Generation
**Issue:** The generated project failed to compile (missing imports) and crash on startup (missing DI wiring).
**Solution:**
- Update the `monolith` template to correctly include `gorm.io/driver/sqlite` when generating `providers.go`.
- Ensure `cmd/server/main.go` correctly exports `*mux.Router` via `fx.Provide` so that dynamically added modules can consume it.

### 2. Consistent Route Prefixes
**Issue:** Core modules use `/api/v1` while added modules default to the root `/`.
**Solution:**
- Update `add module` to check for a global route prefix configuration or detect the existing router setup in `main.go`.
- Allow passing a `--prefix` flag to `add module` (e.g., `kthulu add module customers --prefix=/api/v1`).

### 3. Smart Pluralization
**Issue:** Tables and routes are double-pluralized (e.g., `bookingss`, `customerss`).
**Solution:** Integrate a linguistic pluralization library to handle names correctly (e.g., `bookings` -> `bookings` table, not `bookingss`).

### 4. Relationship Awareness
**Issue:** Foreign keys (`customer_id`, `car_id`) are treated as simple integers without database constraints or ORM relationships.
**Solution:** Support relationship syntax in the CLI, e.g., `kthulu add module bookings car:belongs_to:cars`, to generate proper FK constraints and GORM tags.

### 5. Automatic Auth Integration
**Issue:** New modules are not protected by authentication by default.
**Solution:** Add a `--protected` flag to inject middleware into the generated handler registration logic.

### 6. Missing Migrations & Entry Points
**Issue:** The `users` feature didn't generate a migration, and `cmd/migrate/main.go` was missing despite being in the Makefile.
**Solution:** Ensure all feature flags trigger the necessary file generation, including database migrations and auxiliary command entry points.

By addressing these items, we will achieve a truly "zero-config" experience where the generated code is immediately deployable.
