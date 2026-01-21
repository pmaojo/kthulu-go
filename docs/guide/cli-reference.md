---
title: CLI Reference
description: Complete guide to all Kthulu CLI commands.
---

# CLI Reference

## Project Commands

### `kthulu new` / `kthulu create`

Create a new project.

```bash
kthulu new my-app --features user,auth,product --frontend templ --database postgres
```

**Flags:**

- `--template, -t`: Project template (microservice, monolith, saas, ecommerce, fintech, cli, mcp)
- `--features, -f`: Comma-separated modules to include
- `--database, -d`: Database type (sqlite, postgres, mysql)
- `--frontend`: Frontend type (templ, none)
- `--no-vercel`: Skip Vercel deployment files
- `--from-plan`: Use kthulu-plan.yaml

### `kthulu add module`

Add a module with optional fields.

```bash
kthulu add module orders name:string total:float user:belongs_to:users
```

**Flags:**

- `--protected`: Add auth middleware
- `--admin`: Generate admin CRUD pages
- `--prefix`: Route prefix

### `kthulu doctor`

Check development environment.

```bash
kthulu doctor
```

### `kthulu analyze`

Analyze project architecture.

---

## Database Commands

### `kthulu migrate up`

Run pending migrations.

### `kthulu migrate down`

Rollback last migration.

### `kthulu migrate create`

Create new migration file.

```bash
kthulu migrate create add_orders_table
```

### `kthulu seed`

Run database seeders.

---

## Development Commands

### `kthulu dev`

Start development server with hot-reload.

### `kthulu doc`

Generate OpenAPI documentation.

### `kthulu coder`

Launch AI coding assistant TUI.

---

## Marketplace Commands

### `kthulu marketplace list`

List available modules, starters, plugins.

```bash
kthulu marketplace list --type module
```

### `kthulu marketplace install`

Install a plugin or module.

```bash
kthulu marketplace install aws-deploy
```

---

## Deployment Commands

### `kthulu deploy`

Deploy to cloud providers.

```bash
kthulu deploy --cloud=aws --scale=auto
```

---

## AI Commands

### `kthulu ai`

AI-powered code generation.

```bash
kthulu ai "Add Stripe payments to my API"
```

### `kthulu ai suggest`

Get AI recommendations for your project.

---

## Enterprise Commands

### `kthulu audit`

Run security and compliance checks.

### `kthulu secure`

Audit and patch security vulnerabilities.

---

## Utility Commands

### `kthulu mcp`

Start MCP server for AI agents.

### `kthulu gemini`

Launch Gemini CLI with Kthulu context.

### `kthulu claude`

Configure Claude CLI with Kthulu superpowers.
