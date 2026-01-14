---
title: CLI Reference
description: Comprehensive guide to Kthulu CLI commands.
---


The `kthulu` command line interface is your primary tool for managing projects.

## Core Commands

### `kthulu init`
Initializes the global Kthulu configuration in `~/.kthulu`.

### `kthulu plan`
Starts an interactive session to design your application architecture. Generates a `kthulu-plan.yaml` file.

### `kthulu create`
Scaffolds a new project.
- `--from-plan`: Uses the `kthulu-plan.yaml` in the current directory.
- `--template <name>`: Uses a starter template from the marketplace.

### `kthulu dev`
Starts the development server with hot-reloading and self-healing capabilities.
- `--watch`: Watches for file changes.

### `kthulu add module <name>`
Adds a module from the marketplace or creates a new one.

### `kthulu admin generate <module>`
Generates a React Admin UI for the specified module entity.

### `kthulu deploy`
Deploys the application.
- `--target <provider>`: `wasmer`, `fly`, or `docker`.
