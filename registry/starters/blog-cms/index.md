---
title: "Blog CMS"
description: "Publishing platform: posts, categories and comments with cache and file storage."
type: "starter"
author: "Kthulu Team"
stars: 22
icon: "FileText"
---

# Blog CMS

Publishing platform: posts, categories and comments with cache and file storage.

## Highlights

- Posts with categories, slugs and publish workflow
- Comment moderation via approved flag
- Cache driver for hot pages, storage driver for media uploads
- Admin UI for editors out of the box

## Domain Model

| Entity | Fields |
|--------|--------|
| `category` | name, slug |
| `post` | title, slug, body, published_at, status, topic |
| `comment` | author_name, body, approved, article |

Every entity ships with a typed GORM model, repository, service with
validation, REST API, and a GTH (Templ + HTMX) admin page.

## Get Started

One command — the blueprint ships inside the kthulu binary:

```bash
kthulu marketplace install blog-cms my-app
cd my-app && go run ./cmd/server
```

Or scaffold manually from the plan file at
`registry/starters/blog-cms/plan.yaml`:

```bash
kthulu create my-app --from-plan=registry/starters/blog-cms/plan.yaml
```
