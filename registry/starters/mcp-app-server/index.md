---
title: "MCP App Server"
description: "Model Context Protocol server with an interactive MCP Apps UI (ext-apps) — renders inline in Claude and ChatGPT."
type: "starter"
author: "Kthulu Team"
stars: 47
icon: "Bot"
---

# MCP App Server

Scaffold a complete, dependency-free MCP server that ships an **interactive
UI** using the official [MCP Apps extension](https://github.com/modelcontextprotocol/ext-apps)
(`io.modelcontextprotocol/ui`). Compliant hosts — Claude, ChatGPT and others —
render the view inline in the conversation.

```bash
kthulu create my-mcp-app --template=mcp
cd my-mcp-app && go test ./... && go build ./cmd/my-mcp-app
```

## What you get

- **Zero dependencies**: stdlib-only JSON-RPC 2.0 server over stdio
- **MCP Apps wired end to end**:
  - `ui://` resource served as `text/html;profile=mcp-app`
  - tools linked to views via `_meta.ui.resourceUri`
  - extension capability declared in `initialize`
- **Working example app**: a live status dashboard view that performs the
  `ui/initialize` handshake, renders `ui/notifications/tool-result`
  payloads and calls tools back via `tools/call`
- **Protocol tests included**: initialize/tools/resources round-trips run
  with plain `go test`

## Connect to Claude Desktop

```json
{
  "mcpServers": {
    "my-mcp-app": { "command": "/path/to/my-mcp-app" }
  }
}
```

Ask Claude to "show the status dashboard" and the app renders inline.
