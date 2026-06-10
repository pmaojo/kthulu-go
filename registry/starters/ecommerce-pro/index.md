---
title: "E-commerce Pro"
description: "Storefront backbone: catalog, orders and payments with mail, storage and background jobs."
type: "starter"
author: "Kthulu Team"
stars: 54
icon: "ShoppingCart"
---

# E-commerce Pro

Storefront backbone: catalog, orders and payments with mail, storage and background jobs.

## Highlights

- Catalog, customers, orders and payments wired with foreign keys
- Order/payment state machines enforced by oneof validation
- Mail driver for receipts, storage driver for product images
- Queue runtime for order processing and stock jobs

## Domain Model

| Entity | Fields |
|--------|--------|
| `product` | name, sku, price, stock, description |
| `customer` | name, email, phone |
| `order` | status, total, placed_at, buyer |
| `payment` | amount, method, paid_at, invoice |

Every entity ships with a typed GORM model, repository, service with
validation, REST API, and a GTH (Templ + HTMX) admin page.

## Get Started

One command — the blueprint ships inside the kthulu binary:

```bash
kthulu marketplace install ecommerce-pro my-app
cd my-app && go run ./cmd/server
```

Or scaffold manually from the plan file at
`registry/starters/ecommerce-pro/plan.yaml`:

```bash
kthulu create my-app --from-plan=registry/starters/ecommerce-pro/plan.yaml
```
