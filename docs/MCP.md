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

Ensure you have the `kthulu` binary built and accessible:

```bash
# From the repository root
go build -o kthulu ./cmd/kthulu/main.go
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
      "command": "/absolute/path/to/kthulu-go/kthulu",
      "args": [
        "mcp",
        "--working-dir", "/absolute/path/to/your/target/project"
      ]
    }
  }
}
```

**Note:** The `--working-dir` argument is crucial. It tells the Kthulu MCP server where to execute commands. If omitted, it defaults to the directory where the process is started (which might be the wrong place).

### Session Working Directory

Agents do not need to restart the server to change directories. Two tools
manage a session-wide working directory used by **all** tools (CLI commands,
filesystem, search, tests, git, database):

- `workdir_get` — show the current working directory, whether it's a kthulu
  project, and its top-level entries. Call this first to orient yourself.
- `workdir_set` — point the session at another directory (e.g. right after
  `create` scaffolds a new project). Relative paths resolve against the
  current working directory; an empty path resets to the server default.

A typical agent flow:

1. `create` a project (returns the generated path)
2. `workdir_set` to that path
3. `add module`, `migrate diff`, `go_test`, `fs_*`, ... all now run inside it

### Generating Complex Apps: scaffold_project

The `scaffold_project` tool is the **preferred** way for agents to create
applications. It takes the domain model as structured data, so entities get
their real fields, validation rules and relations instead of the default
single `name` column:

```json
{
  "name": "tournaments",
  "features": ["auth", "user", "queues"],
  "modules": [
    {"name": "tournament", "fields": ["title:string:required,min=3", "starts_at:time", "status:string:oneof=draft|open|running|finished"]},
    {"name": "team",       "fields": ["name:string:required", "city:string", "wins:int"]},
    {"name": "player",     "fields": ["name:string:required", "email:string:email", "squad:belongs_to:team"]},
    {"name": "match",      "fields": ["played_at:time", "home:belongs_to:team", "away:belongs_to:team"]}
  ]
}
```

Modules without fields are rejected with an explanatory error, so agents are
forced to model the domain. After scaffolding, the session working directory
is switched to the new project automatically.

Three more layers keep weak agents on the golden path:

- **Server instructions**: the MCP initialize response describes the
  model-first workflow, so agents read it before choosing tools.
- **`review_domain_model`**: a deterministic reviewer that critiques a
  proposed model (missing relations, enum fields without `oneof`,
  timestamps not typed `time`, emails without validation, plural names,
  no required fields). It also runs automatically inside
  `scaffold_project` and appends suggestions to the result.
- **`create` guardrail**: under MCP, the raw `create` command refuses to
  generate skeleton modules that would fall back to a single default
  `name` field, and redirects the agent to `scaffold_project`.

The raw `create` and `add_module` tools remain available; `add_module`
accepts the same `name:type[:rules]` field syntax.

### Fast Project Creation (No Timeouts)

When driven through MCP, `create` automatically skips the slow
post-generation steps (`templ generate`, `go mod tidy`, `go test`) that
download dependencies and can exceed client tool timeouts. The response
includes the exact commands to finish setup — run them via the `shell_execute`
or `go_build` tools, which you can invoke stepwise. The same behavior is
available on the CLI with `kthulu create --skip-postgen`.

## Debugging

To debug the MCP communication:

1. Use the **MCP Inspector**:
   ```bash
   npx @modelcontextprotocol/inspector \
     /absolute/path/to/kthulu-go/kthulu \
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

Kthulu scaffolds standalone MCP servers with first-class support for the
**MCP Apps extension** (`io.modelcontextprotocol/ui`): generated servers can
ship interactive HTML views that hosts like Claude and ChatGPT render inline.

To create a new MCP server project:

```bash
kthulu new my-mcp-server --template=mcp
```

This will generate a project with:
- `cmd/my-mcp-server/main.go`: The server entrypoint using `mcp-golang`.
- `internal/mcp/`: A dependency-free JSON-RPC 2.0 MCP server with MCP Apps
  support (ui:// resources, _meta.ui tool links, extension negotiation) and
  protocol tests.
- `internal/tools/`: Sample tools, including a status dashboard tool whose
  results render as an interactive app via `internal/tools/ui/dashboard.html`.
- `go.mod`: Pre-configured dependencies.

You can then extend this project by adding more tools in the `internal/tools` directory and registering them in `main.go`.
