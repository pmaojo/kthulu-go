---
title: Quick Start Tour
description: Build, run, and deploy a full-stack application in 5 minutes.
---


Welcome to the Kthulu Tour. We will build a "Task Manager" application with a React frontend and a Go backend.

## 1. Plan Your App

Kthulu uses a `kthulu-plan.yaml` to define your application architecture. Let's create a new plan.

```bash
mkdir my-task-app
cd my-task-app
kthulu plan
```

Follow the interactive prompts:
- **Project Name**: `task-manager`
- **Frontend**: `React`
- **Database**: `PostgreSQL` (or SQLite for simplicity)
- **Modules**: Select `Auth`, `Users`.

## 2. Scaffold the Code

Generate the project structure based on your plan.

```bash
kthulu create --from-plan
```

This will create a hexagonal architecture project:
- `cmd/server`: The entry point.
- `internal/core`: Domain logic.
- `internal/adapters`: HTTP handlers and DB repositories.
- `frontend/`: The React application.

## 3. Add a New Module

Let's add a `tasks` module using the CLI.

```bash
kthulu add module tasks
```

This registers a new module in `internal/core/tasks` and sets up the wiring.

## 4. Define Your Domain

Open `internal/core/tasks/domain.go` and define your Task entity.

```go
package tasks

type Task struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    IsCompleted bool   `json:"is_completed"`
}
```

## 5. Generate Admin UI

Kthulu can automatically generate an Admin UI for your entities.

```bash
kthulu admin generate tasks
```

This inspects your Go struct and generates a React Admin resource in `frontend/src/modules/tasks`.

## 6. Run the Dev Server

Start the development server with self-healing capabilities.

```bash
kthulu dev
```

Visit `http://localhost:3000` to see your app and `http://localhost:3000/admin` for the admin panel.

## 7. Deployment

Deploy your application to the cloud or Wasmer Edge.

```bash
kthulu deploy --target=wasmer
```

## Conclusion

You've just built a modular, full-stack Go application! Explore the [Project Structure](/docs/guide/project-structure) to understand how it works under the hood.
