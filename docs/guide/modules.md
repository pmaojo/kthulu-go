---
title: Modules
description: Discover, create, and share Kthulu modules.
---


Modules are the building blocks of a Kthulu application. A module encapsulates a specific domain capability, such as Authentication, Billing, or Notifications.

## Anatomy of a Module

A standard module structure:

```text
internal/core/billing/
├── domain.go      # Entities
├── ports.go       # Interfaces
└── service.go     # Business Logic
```

## Marketplace

The [Kthulu Marketplace](/marketplace) hosts a collection of verified modules.

To install a module:

```bash
kthulu add module invoice
```

This commands downloads the module code into `internal/core/invoice` and `internal/adapters/...`, and wires it into your application.

## Creating a Module

You can create a custom module using the CLI:

```bash
kthulu add module my-feature
```

This scaffolds the basic directory structure.

## Module Registry

Modules are defined in the `registry/` directory. Each module includes a `metadata.json` describing its dependencies and configuration.

### Metadata Format

```json
{
  "id": "billing",
  "name": "Billing Module",
  "type": "module",
  "dependencies": ["auth", "users"]
}
```
