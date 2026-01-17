# Kthulu Project Context

This is a Kthulu Framework project. It follows a Modular Monolith architecture.

## Architectural Rules
- **Vertical Slices**: Code is organized by business domain (Modules).
- **Layers**: 
  - `API`: Transport layer (HTTP, gRPC).
  - `Core`: Business logic and use cases.
  - `Store`: Data access and persistence.
- **Dependencies**: Modules should be loosely coupled.

## Available Tools (Kthulu MCP)
You have access to a powerful set of tools via the Kthulu MCP server:
- **Shell**: `host_shell_execute`
- **Git**: `git_status`, `git_diff`, `git_log`
- **Architecture**: Create modules, components, and manage the graph.

Use these tools to verify the state of the project and perform tasks.

## Workflows & Troubleshooting

### Dependency Resolution in Workspaces
If you are working in a monorepo or a folder with a parent `go.work` file (e.g., `kthulu-go`):
- **ALWAYS** export `GOWORK=off` when running `go run` or `go test` inside a generated project.
- **Preferred**: Use `kthulu dev` effectively, as it handles this automatically.
- If dependencies are missing ("package not in std"): run `go mod tidy` then try again with `GOWORK=off`.

### Starting the Server
- **Check Ports**: Before starting the server, check if port 8080 is free: `lsof -i :8080`.
- **Use the CLI**: Prefer running `kthulu dev` which starts both backend and frontend with AI self-healing.
- **Manual Start**: `cd <project> && GOWORK=off go run cmd/server/main.go`

### Common Issues
- **"bind: address already in use"**: Kill the process on port 8080 or use `kthulu dev`.
- **"package ... is not in std"**: You are likely mistakenly using the parent `go.work`. Set `GOWORK=off`.
