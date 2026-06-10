---
title: "CRM Pipeline"
description: "Sales CRM: companies, contacts, deals and activities with an event bus and mail."
type: "starter"
author: "Kthulu Team"
stars: 38
icon: "Users"
---

# CRM Pipeline

Sales CRM: companies, contacts, deals and activities with an event bus and mail.

## Highlights

- Classic pipeline domain: companies → leads → deals → activities
- Deal stages constrained to a real sales funnel
- Typed pub/sub event bus to react to stage changes
- Mail driver ready for follow-up sequences

## Domain Model

| Entity | Fields |
|--------|--------|
| `company` | name, domain, industry |
| `lead` | name, email, phone, employer |
| `deal` | title, value, stage, closes_at, account |
| `activity` | kind, notes, happened_at, deal_ref |

Every entity ships with a typed GORM model, repository, service with
validation, REST API, and a GTH (Templ + HTMX) admin page.

## Get Started

```bash
# Copy the plan from this page into my-app-plan.yaml, then:
kthulu create my-app --from-plan=my-app-plan.yaml
cd my-app && go run ./cmd/server
```

The full plan file ships with the framework at
`registry/starters/crm-pipeline/plan.yaml`.
