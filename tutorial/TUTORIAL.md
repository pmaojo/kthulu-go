# Building a Gym Management App with Kthulu CLI

This tutorial will guide you through creating a Gym Management Application using the Kthulu CLI.

## Prerequisites

- Go 1.24+
- Kthulu CLI installed (`kthulu-cli` in your path)

## Step 1: Create the Project

We will start by creating a new project named `gym-app`. We'll use SQLite for the database and include JWT authentication and observability features. We will also pre-configure some modules like users, memberships, classes, bookings, and payments.

```bash
kthulu create gym-app \
  --database sqlite \
  --auth jwt \
  --observability \
  --features users,memberships,classes,bookings,payments
```

This command sets up the project structure, initializes Go modules, and generates the boilerplate code for the specified features.

## Step 2: Explore the Project Structure

Navigate into the `gym-app` directory. The project follows a Clean Architecture structure:

- `cmd/server`: The entry point of the application.
- `internal/adapters/http/modules`: Contains the generated modules (users, memberships, etc.).
- `internal/core`: Core logic and providers.
- `configs`: Configuration files.

Each module (e.g., `internal/adapters/http/modules/users`) has its own `domain`, `repository`, `service`, and `handlers` packages.

## Step 3: Add a New Module

Now, let's add an `equipment` module to manage gym equipment. We will define fields for `name` (string), `quantity` (int), and `last_maintained` (time).

```bash
cd gym-app
kthulu add module equipment name:string quantity:int last_maintained:time
```

This command generates the `equipment` module with:
- Domain entity with the specified fields.
- Repository implementation (GORM).
- Service layer.
- HTTP Handlers.
- Database migration.

**Note:** If the generated module is placed in `internal/modules/equipment` but the project structure uses `internal/adapters/http/modules/`, you may need to move it to align with the rest of the modules.

```bash
mkdir -p internal/adapters/http/modules
mv internal/modules/equipment internal/adapters/http/modules/
```

And update the import in `cmd/server/main.go` if necessary.

## Step 4: Run the Application

Now we can build and run the application.

```bash
go mod tidy
go run cmd/server/main.go
```

The server should start on port 8080. You can access the API endpoints like:
- `GET /users`
- `GET /equipments` (once you have added some data)

## Conclusion

You have successfully created a Gym Management App with Kthulu CLI, added essential modules, and extended it with a custom module.

## Roadmap & Improvements for Kthulu CLI

During the creation of this tutorial, several areas for improvement in the Kthulu CLI were identified to enhance the developer experience:

1.  **Context-Aware Module Generation**
    - **Issue:** `kthulu add module` places new modules in `internal/modules` by default, whereas `kthulu create` uses `internal/adapters/http/modules`.
    - **Improvement:** The CLI should detect the existing project structure and place new modules in the correct location automatically.

2.  **Automated Import & Wiring Resolution**
    - **Issue:** Adding a module required manual updates to `cmd/server/main.go` to fix import paths and register the new module's routes within the `fx` container.
    - **Improvement:** Enhance AST manipulation to robustly update `main.go`, handling import aliases and `fx.Invoke` signature updates seamlessly.

3.  **Self-Contained Module Templates**
    - **Issue:** Generated `module.go` files sometimes miss necessary imports for their own subpackages (e.g., `repository`, `service`).
    - **Improvement:** Ensure generated module entry points correctly import and reference their internal packages.

4.  **Standardized Flag naming**
    - **Issue:** CLI flags for interactivity and automation should be consistent across all commands.
    - **Improvement:** Standardize flags like `--non-interactive` or `--yes` to streamline CI/CD and script usage.
