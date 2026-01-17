# Kthulu Go — AI-Powered Software Foundry

**Kthulu Go** is an intelligent software foundry powered by **Go**, **MCP (Model Context Protocol)**, and **AI**. It is not just a framework, but a Generation Engine designed to autonomously plan, scaffold, and evolve modular software architectures.

The platform prioritizes:

1.  **The Generation Engine**: A powerful CLI that uses AI to translate natural language into robust, type-safe Go code.
2.  **Modular Monolith Architecture**: Enforcing "Vertical Slice" architecture for high cohesion and scalability.
3.  **MCP Native**: Exposing all CLI capabilities as MCP tools, allowing AI agents (like Claude or Gemini) to act as autonomous engineers.

## Who Is Kthulu For?

- **Platform Engineers** building internal developer platforms (IDPs) with strict architectural standards.
- **AI-Native Teams** who want to leverage agents to drastically reduce boilerplate and maintenance overhead.
- **Go Developers** seeking a modern, opinionated framework that balances simplicity with enterprise readiness.

## Intelligent Coding (TUI)

Kthulu replaces complex internal AI implementations with **high-leverage integrations**.

```sh
# Autoconfigures & Launches "Crush" with Kthulu tools
kthulu coder
```

When you run `kthulu coder`, it:

1. Detects your installation of [Crush](https://github.com/charmbracelet/crush).
2. Generates a project-specific configuration that **injects Kthulu as an MCP Server**.
3. Launches the agent, giving it full access to scaffold, analyze, and evolve your code.

## Model Context Protocol (MCP)

Kthulu Go is built first and foremost as an **MCP Server**. This means it is designed to be driven by AI.

To use Kthulu with other agents, you can use our bridge commands:

```sh
# Configure & Launch Claude CLI
kthulu claude

# Configure Gemini Code Assist
kthulu gemini
```

Or manually register the server in your favorite client:

```json
{
  "mcpServers": {
    "kthulu": {
      "command": "/absolute/path/to/kthulu",
      "args": ["mcp"]
    }
  }
}
```

## Extensions & Ecosystem

Kthulu is designed to be extensible. We support two primary ways to extend capabilities:

### 1. Kthulu Modules (Vertical Slices)

Native Go plugins that live inside your `internal/modules` directory. They add business logic and API endpoints to your application.

### 2. Agent Skills (MCP)

You can extend the AI's capabilities by registering additional MCP servers (Extensions).

- **Database Skills**: Grants the agent SQL access to your database.
- **Git Skills**: Allows the agent to open PRs and manage branches.
- **Browser Skills**: Lets the agent research documentation online.

When using `kthulu coder` (Crush), you can manage these extensions in `.kthulu/crush.json`.

## Key Features

- **Vertical Slice Architecture**: Code is organized by feature, not technical layer.
- **Dependency Injection**: First-class support for `uber/fx`.
- **Zero-Boilerplate**: The CLI handles wiring, config, and scaffolding.
- **Database Agnostic**: Built-in support for SQLite, PostgreSQL, and MySQL via GORM.

## Project Structure

This repository is a **Monorepo** containing the seed of the foundry:

```
kthulu-go/
├── cmd/
│   └── kthulu/          # The Kthulu CLI & Generation Engine
├── pkg/                 # Shared libraries and public APIs
├── internal/            # Private framework internals
└── verify-v*/           # Generated verification apps (ephemeral)
```

## Getting Started

### Installation

You can install the `kthulu` binary directly:

```sh
# Build from source
go build -o kthulu ./cmd/kthulu/main.go

# Add to PATH
export PATH="$(pwd):$PATH"
```

### Creating a New Project

The core value of Kthulu is its ability to generate production-ready code.

```sh
# Create a new Modular Monolith project
kthulu create my-app
```

This generates a project structure optimized for vertical slicing:

```
my-app/
├── cmd/server/          # Entrypoint
├── internal/
│   ├── modules/         # Vertical Slices (Feature Modules)
│   │   ├── user/        # 'User' Feature
│   │   │   ├── api/     # HTTP Handlers / Transport
│   │   │   ├── core/    # Domain Logic & Services
│   │   │   ├── store/   # Data Persistence
│   │   │   └── module.go # FX Dependency Injection
│   │   └── ...
│   └── infrastructure/  # Shared tech (Loggers, Middleware)
└── ...
```

## CLI Command Reference

| Command                       | Description                                              |
| ----------------------------- | -------------------------------------------------------- |
| `kthulu create <name>`        | Scaffolds a new project with Modular Monolith structure. |
| `kthulu coder`                | Launches the AI Coding Assistant (Crush).                |
| `kthulu claude`               | Configures and launches Claude CLI.                      |
| `kthulu gemini`               | Configures Gemini Code Assist.                           |
| `kthulu dev`                  | Starts the development server with **AI Self-Healing**.  |
| `kthulu add module <name>`    | Adds a new feature module (Vertical Slice).              |
| `kthulu add component <type>` | Adds a component (handler, service, store) to a module.  |
| `kthulu doc`                  | Generates OpenAPI/Swagger documentation.                 |
| `kthulu secure`               | Audits dependencies for vulnerabilities.                 |
| `kthulu audit`                | Runs enterprise compliance checks.                       |
| `kthulu analyze`              | Analyzes project structure and dependencies.             |

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

MIT — see [LICENSE](./LICENSE)
