# Kthulu Go MCP Server

The Kthulu Go MCP (Model Context Protocol) server provides a powerful interface for AI agents to interact with the Kthulu Go CLI. This allows agents to automate a wide range of Go development tasks, from project creation to code generation and security scanning.

## Starting the Server

To start the MCP server, run the following command from the `backend/backend` directory:

```bash
go run ./cmd/kthulu-cli mcp
```

By default, the server will use the current directory as its working directory. You can specify a different working directory using the `--working-dir` flag:

```bash
go run ./cmd/kthulu-cli mcp --working-dir /path/to/your/project
```

If the server is started within a Go project (i.e., a directory containing a `go.mod` file), it will automatically set its working directory to the project's root.

## Available Tools

The MCP server exposes the following tools for AI agents to use:

*   **`create_project`**: Creates a new Go project using the Kthulu CLI.
    *   **Arguments:**
        *   `name` (string, required): The name of the project to create.
        *   `template` (string, optional): The template to use for the new project.
    *   **Behavior:** When this tool is used, the server's working directory will automatically switch to the newly created project's directory.

*   **`guide_tagging`**: Analyzes the project and guides the user on adding Kthulu tags.
    *   **Behavior:** This tool will scan the project for untagged Go files and return a list of suggestions for adding Kthulu tags.

*   **`add_module`**: Adds a new module to the project.
    *   **Arguments:**
        *   `name` (string, required): The name of the module to add.

*   **`add_component`**: Adds a new component to the project.
    *   **Arguments:**
        *   `type` (string, required): The type of the component to add (e.g., handler, service, repository).
        *   `name` (string, required): The name of the component to add.

*   **`generate_code`**: Generates code artifacts.
    *   **Arguments:**
        *   `type` (string, required): The type of code to generate (e.g., handler, usecase, entity).
        *   `name` (string, required): The name of the artifact to generate.

*   **`run_ai_assistant`**: Invokes the AI assistant.
    *   **Arguments:**
        *   `prompt` (string, required): The prompt for the AI assistant.

*   **`manage_database`**: Manages database migrations.
    *   **Arguments:**
        *   `subcommand` (string, required): The migration subcommand to run (e.g., up, down, status).

*   **`security_scan`**: Scans for security vulnerabilities.

*   **`project_management`**: Manages the project.
    *   **Arguments:**
        *   `command` (string, required): The project management command to run (e.g., audit, deploy, status).
