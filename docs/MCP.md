# Kthulu Model Context Protocol (MCP) Server

Kthulu exposes its CLI capabilities as an MCP Server, allowing AI agents (like Claude or custom agents) to interact with your project structure, generate code, and perform maintenance tasks.

## Features

When running as an MCP server, Kthulu exposes the following **Tools**:

- **create**: Scaffold a new project using the `kthulu create` logic.
- **add_module**: Add a new module to an existing project (`kthulu add module`).
- **add_component**: Add specific components (handlers, services, etc.).
- **generate**: Generate code artifacts.
- **ai**: Invoke the internal AI assistant for code suggestions.
- **secure**: Run security scans.
- **migrate**: Manage database migrations.
- **inspect**: (If available) Read project structure and configuration.

## Installation & Configuration

### 1. Build the Binary

Ensure you have the `kthulu-cli` binary built and accessible:

```bash
cd backend/backend
go build -o ../../bin/kthulu-cli ./cmd/kthulu-cli
```

### 2. Configure Claude Desktop

Edit your configuration file:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%/Claude/claude_desktop_config.json`

Add the server configuration:

```json
{
  "mcpServers": {
    "kthulu": {
      "command": "/absolute/path/to/kthulu-go/bin/kthulu-cli",
      "args": [
        "mcp",
        "--working-dir", "/absolute/path/to/your/target/project"
      ]
    }
  }
}
```

**Note:** The `--working-dir` argument is crucial. It tells the Kthulu MCP server where to execute commands. If omitted, it defaults to the directory where the process is started (which might be the wrong place).

## Debugging

To debug the MCP communication:

1. Use the **MCP Inspector**:
   ```bash
   npx @modelcontextprotocol/inspector \
     /absolute/path/to/kthulu-go/bin/kthulu-cli \
     mcp --working-dir /path/to/test/project
   ```

2. Check Logs:
   The Kthulu MCP server writes logs to `stderr`. You can redirect this to a file in your config if your client supports it, or use the inspector to view logs in real-time.

## Publishing to Registries

If you wish to publish this server to the [community list](https://github.com/modelcontextprotocol/servers):

1. Ensure the repository is public.
2. Create a `server.json` (or equivalent metadata) if required by the registry.
3. Submit a Pull Request to the registry repository adding your server details.

Since Kthulu is a CLI tool rather than a hosted API, it is primarily distributed via binary/source and run locally by the user ("stdio" transport).

## Scaffolding New MCP Servers

Kthulu can now scaffold new, standalone MCP servers to help you build your own tools for AI agents.

To create a new MCP server project:

```bash
kthulu new my-mcp-server --template=mcp
```

This will generate a project with:
- `cmd/my-mcp-server/main.go`: The server entrypoint using `mcp-golang`.
- `internal/tools/`: A sample tool implementation.
- `go.mod`: Pre-configured dependencies.

You can then extend this project by adding more tools in the `internal/tools` directory and registering them in `main.go`.
