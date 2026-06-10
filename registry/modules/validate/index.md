---
title: "Validate"
description: "Declarative validation rules in the field DSL: generated Validate() methods, service-level enforcement and 422 responses."
type: "module"
author: "Kthulu Core"
stars: 0
icon: "Box"
---

# Validate

Declare validation rules directly in your blueprint or CLI field definitions using the `name:type:rules` syntax. Kthulu generates a plain-Go `Validate()` method per entity — no reflection, no struct-tag magic — enforced automatically by the service layer.

```yaml
modules:
  product:
    fields:
      - name:string:required,min=2
      - price:int:required,min=1
      - status:string:oneof=draft|active|archived
      - contact_email:string:email
```

## Supported rules

| Rule | Applies to | Behavior |
|------|-----------|----------|
| `required` | string, numbers, time | non-empty / non-zero |
| `min=N` / `max=N` | string, numbers | length for strings, value for numbers |
| `email` | string | RFC-style email shape |
| `oneof=a\|b\|c` | string | enumerated values |

## What gets generated

- `core/<module>_validation.go` — a `ValidationErrors` type and a readable `Validate()` method built from your rules.
- Service layer calls `entity.Validate()` on create and update.
- HTTP handlers translate validation failures into **422 Unprocessable Entity** with a field-level error map:

```json
{
  "errors": {
    "name": "must be at least 2 characters",
    "price": "must be at least 1",
    "status": "must be one of: draft, active, archived"
  }
}
```

Every module gets a compiling `Validate()` even without declared rules, so service code never breaks.
