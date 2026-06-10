---
title: Modules
description: Explore and use the 35+ modules in Kthulu.
---

# Modules

Modules are the building blocks of a Kthulu application. Each encapsulates a specific domain capability.

## Available Modules (35+)

### Infrastructure

| Module      | Description                                          |
| ----------- | ---------------------------------------------------- |
| `mail`      | Mailer driver (SMTP + log; provider SDK stubs)       |
| `cache`     | Caching (Memory, Redis, Memcached)                   |
| `storage`   | File storage (local driver; S3/GCS/Azure stubs)      |
| `queues`    | Background jobs: retries, dead letters, scheduler    |
| `events`    | Pub/Sub event system                                 |
| `policy`    | Authorization gates & policies                       |
| `rate`      | Rate limiting (Token Bucket, Sliding Window)         |
| `session`   | Session management                                   |
| `i18n`      | Internationalization                                 |
| `validate`  | Declarative field rules -> Validate() + 422 errors   |
| `seeder`    | Database seeding + Faker                             |

### Business

| Module         | Description                |
| -------------- | -------------------------- |
| `user`         | User management & profiles |
| `auth`         | JWT authentication + RBAC  |
| `product`      | Product catalog            |
| `invoice`      | Invoicing & billing        |
| `inventory`    | Stock management           |
| `organization` | Multi-tenant orgs          |
| `contact`      | CRM integration            |
| `calendar`     | Scheduling                 |
| `notification` | Push/Email/SMS             |
| `order`        | Order management           |
| `payments`     | Payment processing         |

### Enterprise

| Module          | Description               |
| --------------- | ------------------------- |
| `oauthsso`      | OAuth 2.0 / SSO           |
| `audit`         | Audit logging             |
| `realtime`      | WebSockets                |
| `observability` | Metrics, tracing, logging |
| `verifactu`     | Spanish tax compliance    |
| `flags`         | Feature flags             |
| `health`        | Health checks             |

---

## Installing a Module

```bash
# During project creation
kthulu new my-app --features mail,cache,storage

# Add to existing project
kthulu add module scheduler
```

---

## Module Structure

```text
internal/modules/mail/
├── api/           # HTTP handlers
│   └── handler.go
├── core/          # Business logic
│   ├── entity.go
│   └── service.go
└── store/         # Data access
    └── repository.go
```

---

## Environment Variables

Each infrastructure module uses environment variables:

```env
# Mail
MAIL_DRIVER=smtp
MAIL_HOST=smtp.example.com

# Cache
CACHE_DRIVER=redis
REDIS_HOST=localhost

# Storage
STORAGE_DRIVER=s3
S3_BUCKET=my-bucket

# And more...
```

---

## Browse the Marketplace

Explore all modules at [/marketplace](/marketplace).
