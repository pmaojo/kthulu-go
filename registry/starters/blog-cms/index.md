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

```bash
# Copy the plan from this page into my-app-plan.yaml, then:
kthulu create my-app --from-plan=my-app-plan.yaml
cd my-app && go run ./cmd/server
```

The full plan file ships with the framework at
`registry/starters/blog-cms/plan.yaml`.
