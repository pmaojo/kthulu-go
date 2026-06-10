---
title: "Storage"
description: "StorageModule provides file storage functionality. Supports multiple drivers: local, s3, gcs, azure"
type: "module"
author: "Kthulu Core"
stars: 0
icon: "Box"
---

# Storage

StorageModule provides file storage functionality. Supports multiple drivers: local, s3, gcs, azure

## Features


- Auto-configured Fx Module
- Clean Architecture structure



## Configuration

The module is configured via environment variables:

| Variable | Description |
|----------|-------------|
| - | No environment variables detected |


## Installation

Add this module to your project:

```bash
kthulu add module storage
```

## Components

This module provides the following components to the application:

- **Backend**: Go module wired with Uber Fx.


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






