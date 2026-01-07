# Tutorial: Building a Car Farm Management App with Kthulu CLI

This tutorial describes the process of creating a "Car Farm Management" application using the `kthulu-cli`. The application is designed to manage cars, farms, and maintenance records.

## Prerequisites

- **Go**: Ensure Go 1.24+ is installed.
- **Kthulu CLI**: The CLI must be built and available.

## Steps

### 1. Build the CLI

First, we need to build the `kthulu-cli` tool from the source code.

```bash
cd backend/backend
go build -o ../../bin/kthulu-cli ./cmd/kthulu-cli
```

### 2. Create the Project

We create a new project named `car-farm-app` using the `microservice` template (default) with SQLite database and observability features enabled.

```bash
bin/kthulu-cli create car-farm-app --database sqlite --observability
```

This command generates the project structure, including:
- `cmd/server/main.go`: The entry point.
- `internal/`: Core logic and modules.
- `go.mod`: Dependency management.
- `Dockerfile`, `Makefile`: Build and deployment configurations.

### 3. Add Core Modules

We add three modules to the application: `cars`, `farms`, and `maintenance`.

#### Add `cars` module
The `cars` module tracks vehicle details.

```bash
cd car-farm-app
../bin/kthulu-cli add module cars make:string model:string year:int vin:string status:string -y
```

#### Add `farms` module
The `farms` module tracks farm locations and capacity.

```bash
../bin/kthulu-cli add module farms name:string location:string capacity:int -y
```

#### Add `maintenance` module
The `maintenance` module tracks repairs and costs, linked to a specific car. Note the relationship syntax `car:belongs_to:cars`.

```bash
../bin/kthulu-cli add module maintenance description:string cost:float date:time car:belongs_to:cars -y
```

### 4. Verify the Project

Run tests to ensure everything is wired correctly.

```bash
go mod tidy
go test ./...
```

Run the application:

```bash
go run cmd/server/main.go
```

### 5. Generate Documentation

Generate Swagger API documentation.

```bash
../bin/kthulu-cli doc
```

## Roadmap & Frictions

During the process, several frictions were encountered. Below is a roadmap to address them.

### Frictions Encountered

1.  **Broken Test in Template**: The generated `cmd/server/main_test.go` called `setupRoutes()`, but the actual function in `main.go` was named `NewRouter()`. This required manual intervention to fix before tests could pass.
2.  **Incorrect Import in Generated Module**: When adding the `maintenance` module with a relationship to `cars`, the generated import path in `domain/maintenance.go` was malformed (`"carsDomain "//cars/domain""`). It should have been a valid Go import path.
3.  **Type Name Mismatch**: The generated code assumed the struct name for the `cars` module was `Car` (singular), but it was generated as `Cars` (plural/module name). This caused compilation errors in the `maintenance` module which referenced `carsDomain.Car`.
4.  **CLI Flag Confusion**: The `--fields` flag mentioned in some contexts (or assumed) does not exist; fields are passed as positional arguments.
5.  **Interactive Mode EOF**: Running the CLI in a non-interactive environment without `-y` caused an EOF error on prompt, which is expected but worth noting for automation scripts.
6.  **`swag` Tool Dependency**: `kthulu-cli doc` failed initially because `swag` was not installed, though it attempted to install it.

### Roadmap to Fix

1.  **Fix Project Template**:
    - Update `templates/backend/cmd/server/main_test.go.tmpl` (or equivalent) to use the correct function name `NewRouter` instead of `setupRoutes`.

2.  **Improve Import Generation**:
    - Review the `add module` logic in `backend/backend/cmd/kthulu-cli/cmd/add.go` (or `internal/generator`) to ensure it generates valid, absolute import paths for cross-module dependencies, avoiding double quotes or comments in the import string.

3.  **Standardize Naming Conventions**:
    - Ensure the generator correctly singularizes struct names (e.g., module `cars` -> struct `Car`, table `cars`) or consistently uses the module name. The `inflection` library usage should be verified.
    - Fix the relationship generation to reference the correct struct name (e.g., check if the target module's struct is `Car` or `Cars`).

4.  **Enhance CLI UX**:
    - Update help text to clearly show usage of positional arguments for fields vs flags.
    - Improve error handling for interactive prompts in non-interactive shells.

5.  **Robust Tool Management**:
    - Ensure `kthulu-cli doc` checks for `swag` in `GOPATH/bin` correctly and handles installation failures gracefully, or instructs the user clearly.
