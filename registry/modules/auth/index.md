---
title: "Auth Module"
description: "Production-ready authentication with JWT, OAuth2, and RBAC support."
type: "module"
author: "Kthulu Core"
stars: 100
icon: "Shield"
---

# Auth Module

The Auth module provides a complete authentication solution for your Kthulu application. It supports JWT-based stateless authentication, OAuth2 providers (Google, GitHub), and Role-Based Access Control (RBAC).

## Features

- **JWT Authentication**: Secure, stateless authentication with access and refresh tokens.
- **OAuth2 Support**: Built-in support for Google, GitHub, and custom providers.
- **RBAC**: Define roles and permissions to granularly control access to resources.
- **Middleware**: Ready-to-use HTTP middleware for protecting routes.
- **Database Agnostic**: Works with PostgreSQL, MySQL, and SQLite.

## Installation

Add the module to your project using the CLI:

```bash
kthulu add module auth
```

This will:
1.  Scaffold the `internal/core/auth` domain logic.
2.  Add `AuthHandler` to your HTTP adapter.
3.  Inject the `AuthMiddleware` into your router.

## Configuration

The module is configured via environment variables.

| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_SECRET` | Secret key for signing tokens | `changeme` |
| `JWT_EXPIRY` | Token expiration time | `24h` |
| `OAUTH_GOOGLE_ID` | Google Client ID | - |
| `OAUTH_GOOGLE_SECRET` | Google Client Secret | - |

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
