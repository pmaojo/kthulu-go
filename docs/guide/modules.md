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

## Documenting Modules

To enrich the documentation page for a module (displayed in the Hub/Marketplace), you can create or edit the `index.md` file within the module's registry folder.

Location: `registry/modules/<module-name>/index.md`

### Frontmatter

The documentation uses Frontmatter to define metadata used by the Hub UI.

```yaml
---
title: "Billing Module"
description: "Handle subscriptions, invoices, and payments."
type: "module"
author: "Kthulu Team"
stars: 50
icon: "CreditCard"
---
```

### Content

The content after the frontmatter is rendered as Markdown. You should include:

- **Features**: A bulleted list of what the module does.
- **Installation**: Specific installation instructions if any.
- **Configuration**: Environment variables or config structs.
- **Usage**: Code examples for the Domain or Service layer.
- **API Reference**: HTTP/gRPC endpoints exposed.
