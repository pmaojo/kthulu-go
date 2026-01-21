---
title: Quick Start Tour
description: Build, run, and deploy a full-stack application in 5 minutes with all Kthulu features.
---

# Quick Start Tour

Build a production-ready Go application with **35+ modules** in minutes.

## Prerequisites

```bash
kthulu doctor
```

✅ Go 1.21+ ✅ Git ✅ Node.js

---

## 1. Create Your Project

```bash
# Simple start
kthulu new my-app

# With specific modules
kthulu new my-shop --features user,auth,product,invoice

# Full-stack with infrastructure modules
kthulu new saas-app \
  --features user,auth,mail,cache,storage,scheduler,events \
  --frontend templ \
  --database postgres
```

**Templates:** `microservice`, `monolith`, `saas`, `ecommerce`, `fintech`, `cli`, `mcp`

---

## 2. Run Migrations

```bash
cd my-app
kthulu migrate up
```

Output:

```
✅ create_user_table.sql
✅ create_auth_table.sql
✅ create_product_table.sql
...
goose: successfully migrated to version: 20260121001233
```

---

## 3. Start Development

```bash
kthulu dev
```

Visit [http://localhost:8080](http://localhost:8080)

- 🔧 Hot-reload backend
- ⚡ Templ/HTMX frontend
- 🤖 AI self-healing

---

## Available Modules (35+)

### Infrastructure

| Module      | Description                  |
| ----------- | ---------------------------- |
| `mail`      | SMTP, SES, SendGrid, Mailgun |
| `cache`     | Memory, Redis, Memcached     |
| `storage`   | Local, S3, GCS, Azure        |
| `scheduler` | Cron-like task scheduling    |
| `events`    | Pub/Sub event system         |
| `policy`    | Authorization gates          |
| `rate`      | Rate limiting                |
| `session`   | Session management           |
| `i18n`      | Internationalization         |
| `validate`  | Advanced validation          |
| `seeder`    | Database seeding with Faker  |

### Business

| Module         | Description                |
| -------------- | -------------------------- |
| `user`         | User management            |
| `auth`         | JWT + OAuth authentication |
| `product`      | Product catalog            |
| `invoice`      | Invoicing system           |
| `inventory`    | Stock management           |
| `calendar`     | Scheduling                 |
| `notification` | Push/Email/SMS             |

### Enterprise

| Module      | Description            |
| ----------- | ---------------------- |
| `oauthsso`  | OAuth 2.0 / SSO        |
| `audit`     | Audit logging          |
| `realtime`  | WebSockets             |
| `verifactu` | Spanish tax compliance |

---

## CLI Commands

```bash
# Project
kthulu new my-app --features user,auth
kthulu add module orders name:string total:float
kthulu doctor
kthulu analyze

# Database
kthulu migrate up
kthulu migrate create add_orders_table
kthulu seed

# Development
kthulu dev
kthulu doc

# Marketplace
kthulu marketplace list
kthulu marketplace install aws-deploy

# Deployment
vercel deploy
kthulu deploy --cloud=aws
```

---

## Environment Variables

```env
# Mail
MAIL_DRIVER=smtp           # smtp, ses, sendgrid
MAIL_HOST=smtp.example.com

# Cache
CACHE_DRIVER=memory        # memory, redis, memcached
REDIS_HOST=localhost

# Storage
STORAGE_DRIVER=local       # local, s3, gcs, azure
S3_BUCKET=my-bucket

# Session
SESSION_DRIVER=memory      # memory, redis, database
SESSION_LIFETIME=2h

# i18n
APP_LOCALE=en
TRANSLATIONS_DIR=./translations
```

---

## Project Structure

```
my-app/
├── cmd/server/main.go
├── configs/app.yaml
├── internal/
│   ├── modules/
│   │   ├── user/
│   │   │   ├── api/      # HTTP handlers
│   │   │   ├── core/     # Business logic
│   │   │   └── store/    # Data access
│   │   └── auth/
│   ├── views/            # Templ templates
│   └── infrastructure/
├── migrations/
└── pkg/bootstrap/
```

---

## Next Steps

1. **Add modules:** `kthulu add module reviews rating:int --protected`
2. **Generate docs:** `kthulu doc`
3. **Deploy:** `vercel deploy --prod`
4. **Explore marketplace:** `kthulu marketplace list`

→ [Project Structure](/docs/guide/project-structure) | [Modules](/docs/guide/modules) | [CLI Reference](/docs/guide/cli-reference)
