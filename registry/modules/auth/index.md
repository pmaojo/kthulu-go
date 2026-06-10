---
title: "Auth"
description: "AuthModule provides authentication functionality. Repositories are injected via the ModuleSet provider map to avoid duplication."
type: "module"
author: "Kthulu Core"
stars: 100
icon: "Shield"
---

# Auth

AuthModule provides authentication functionality. Repositories are injected via the ModuleSet provider map to avoid duplication.

## Features

- HTTP API
- Domain Logic



## Configuration

The module is configured via environment variables:

| Variable | Description |
|----------|-------------|
| - | No environment variables detected |


## Installation

Add this module to your project:

```bash
kthulu add module auth
```

## Components

This module provides the following components to the application:

- **Backend**: Go module wired with Uber Fx.


## Usage

### protecting Routes

To protect a route, wrap it with the `AuthMiddleware`.

```go
// internal/adapters/http/router.go

func (r *Router) RegisterRoutes(auth *auth.Middleware) {
    // Public routes
    r.GET("/login", authHandler.Login)

    // Protected routes
    api := r.Group("/api")
    api.Use(auth.RequireAuth)
    {
        api.GET("/profile", userHandler.GetProfile)
    }

    // Admin only
    admin := r.Group("/admin")
    admin.Use(auth.RequireRole("admin"))
    {
        admin.DELETE("/users/:id", userHandler.DeleteUser)
    }
}
```

### Retrieving User Context

You can retrieve the authenticated user from the context in your handlers.

```go
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
    user, ok := auth.FromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    json.NewEncoder(w).Encode(user)
}
```

## API Reference

### `POST /auth/login`

Authenticates a user and returns a token pair.

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:**

```json
{
  "access_token": "eyJhbG...",
  "refresh_token": "dGhpc...",
  "expires_in": 3600
}
```

## Troubleshooting

- **Token Invalid**: Ensure the `JWT_SECRET` matches between the issuer and the verifier.
- **CORS Errors**: Check your CORS configuration in `internal/infrastructure/config/config.go`.













