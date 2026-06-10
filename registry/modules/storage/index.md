---
title: "Storage"
description: "Generated file storage infrastructure: a Storage interface with a complete local-disk driver, env-configured, injected via Fx."
type: "module"
author: "Kthulu Core"
stars: 0
icon: "Box"
---

# Storage

Add `storage` to your project features and Kthulu generates `internal/infrastructure/storage/` — a `Storage` interface with a complete **local disk** driver (put/get/list/copy/move/delete, URLs, metadata) plus generated tests. The driver is provided through Fx, so any service can inject it.

```bash
kthulu create my-app --features=auth,user,storage
```

## Usage

```go
func NewUploadService(store storage.Storage) *UploadService { ... }

store.PutBytes(ctx, "avatars/42.png", data)
data, _ := store.GetBytes(ctx, "avatars/42.png")
files, _ := store.List(ctx, "avatars")
url := store.URL("avatars/42.png")
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `STORAGE_DRIVER` | `local` | `local` (S3/GCS/Azure are scaffolded stubs) |
| `STORAGE_ROOT` | `./storage` | Base path for local files |
| `STORAGE_URL` | `/storage` | Base URL for public files |

The cloud drivers (S3, GCS, Azure) are generated as typed stubs with clear errors pointing at the SDK dependency to add.
