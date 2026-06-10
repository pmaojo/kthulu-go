---
title: Queues Module
description: Database-backed background jobs with retries, dead letters and a recurring scheduler. No Redis required.
---

# Queues Module

Add `queues` (aliases: `queue`, `jobs`, `scheduler`) to your project features and Kthulu generates a complete background-job runtime backed by your existing database — SQLite, PostgreSQL or MySQL. No Redis or extra infrastructure.

```bash
kthulu create my-app --features=auth,user,queues
```

## What gets generated

| Path | Purpose |
|------|---------|
| `internal/infrastructure/queue/queue.go` | Worker pool, job claiming, retries, scheduler |
| `internal/infrastructure/queue/queue_test.go` | Runtime tests: processing, retries, dead-lettering |
| `internal/jobs/jobs.go` | Your application's job handlers (the `app/Jobs` equivalent) |

The worker pool is wired into the Fx application lifecycle: it starts with the app and drains gracefully on shutdown.

## Defining and enqueuing jobs

```go
// internal/jobs/jobs.go
const TypeWelcomeEmail = "email.welcome"

type WelcomeEmail struct {
    UserID uint   `json:"user_id"`
    Email  string `json:"email"`
}

func Register(q *queue.Queue) error {
    q.Register(TypeWelcomeEmail, HandleWelcomeEmail)
    return q.Every(time.Hour, TypeHeartbeat, nil) // cron equivalent
}
```

```go
// From any service or handler:
q.Enqueue(jobs.TypeWelcomeEmail, jobs.WelcomeEmail{UserID: 42})
q.EnqueueIn(10*time.Minute, jobs.TypeWelcomeEmail, payload)
```

## Semantics

- **At-least-once execution** with optimistic job claiming — concurrent workers never run the same job twice.
- **Retries with exponential backoff** (configurable budget, default 5 attempts).
- **Dead letters**: exhausted jobs stay in the `jobs` table with `last_error` for inspection.
- **Panic recovery**: a panicking handler fails the job instead of crashing the worker.
- **Recurring schedules** declared in code via `q.Every(interval, type, payload)`.

## Configuration

```go
queue.New(db,
    queue.WithConcurrency(8),
    queue.WithPollInterval(500*time.Millisecond),
    queue.WithMaxAttempts(10),
)
```
