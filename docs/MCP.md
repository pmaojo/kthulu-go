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
- **inspect**: Read project structure and configuration.

## Installation & Configuration

### 1. Build the Binary

Ensure you have the `kthulu-cli` binary built and accessible:

```bash
cd backend
go build -o ../bin/kthulu-cli ./cmd/kthulu-cli
```

### 2. Configure Claude Desktop

Edit your configuration file:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%/Claude/claude_desktop_config.json`

Add the server configuration. **Note that you must provide your `GEMINI_API_KEY` in the `env` block** so the CLI allows AI features to work properly.

```json
{
  "mcpServers": {
    "kthulu": {
      "command": "/absolute/path/to/kthulu-go/bin/kthulu-cli",
      "args": [
        "mcp",
        "--working-dir", "/absolute/path/to/your/target/project"
      ],
      "env": {
        "GEMINI_API_KEY": "your-gemini-api-key-here"
      }
    }
  }
}
```

**Note:** The `--working-dir` argument is crucial. It tells the Kthulu MCP server where to execute commands. If omitted, it defaults to the directory where the process is started.

### 3. Manual Usage

If you are running the MCP server manually (e.g. for testing or another client), ensure the environment variable is set:

```bash
export GEMINI_API_KEY=your-gemini-api-key-here
./bin/kthulu-cli mcp --working-dir /path/to/project
```

## Configuration Flags

The `mcp` command supports several flags to customize its behavior:

| Flag | Description | Default |
|------|-------------|---------|
| `--working-dir` | Working directory for executed CLI commands | Current directory |
| `--transport` | Transport for MCP server: `stdio` or `http` | `stdio` |
| `--listen` | Listen address when using the HTTP transport | `:8080` |
| `--http-path` | HTTP path for MCP requests when transport=`http` | `/mcp` |
| `--allow` | Whitelist of CLI command paths (e.g. `migrate up`). Only these commands will be exposed. | None (all exposed) |
| `--deny` | Blacklist of CLI command paths (e.g. `deploy apply`). Denials override allows. | None |

## Debugging

To debug the MCP communication:

1. Use the **MCP Inspector**:
   ```bash
   export GEMINI_API_KEY=your-key
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
