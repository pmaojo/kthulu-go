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
