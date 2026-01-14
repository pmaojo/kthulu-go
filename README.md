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

## Model Context Protocol (MCP)

Kthulu Go is built first and foremost as an **MCP Server**. This means it is designed to be driven by AI.

To use Kthulu with **Claude Desktop** or other MCP clients:

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

Once connected, your AI assistant gains the ability to:

- **Analyze** your codebase structure.
- **Plan** new features and modules.
- **Generate** code that complies with your project's architecture.
- **Verify** integration and dependencies.

## Key Features

- **Vertical Slice Architecture**: Code is organized by feature, not technical layer, making it easy to add, remove, or extract capabilities.
- **Dependency Injection**: First-class support for `uber/fx`.
- **Zero-Boilerplate**: The CLI handles wiring, config, and scaffolding.
- **Database Agnostic**: Built-in support for SQLite, PostgreSQL, and MySQL via GORM.

## CLI Command Reference

| Command | Description |
|Utils|---|
| `kthulu create <name>` | Scaffolds a new project with Modular Monolith structure. |
| `kthulu dev` | Starts the development server with **AI Self-Healing**. |
| `kthulu add module <name>` | Adds a new feature module (Vertical Slice). |
| `kthulu add component <type>` | Adds a component (handler, service, store) to a module. |
| `kthulu doc` | Generates OpenAPI/Swagger documentation. |
| `kthulu secure` | Audits dependencies for vulnerabilities. |
| `kthulu audit` | Runs enterprise compliance checks. |
| `kthulu analyze` | Analyzes project structure and dependencies. |
| `kthulu ai suggest` | Asks the AI for code improvements or refactoring. |

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

MIT — see [LICENSE](./LICENSE)
