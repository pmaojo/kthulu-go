# Standard Library Migration Report

## Executive Summary

This report outlines the strategy to reduce binary size, improve build times, and align with Go's modern standard library by replacing third-party dependencies with standard equivalents.

**Primary Targets:**
*   **Web Frameworks:** `github.com/go-chi/chi` and `github.com/gin-gonic/gin` &rarr; `net/http` (Go 1.22+).
*   **Logging:** `go.uber.org/zap` &rarr; `log/slog` (Go 1.21+).
*   **Testing:** Adopting BDD (Behavior Driven Development) while retaining `github.com/stretchr/testify` for assertions.

## 1. Web Framework Migration (`net/http`)

### Current State
*   **`chi`**: extensively used in `internal/adapters/http`, `cmd/service`, and `kthulu-cli` templates.
*   **`gin`**: Used sparsely (likely legacy or isolated modules like `rbac_enterprise.go`).

### The `net/http` Advantage (Go 1.22+)
Go 1.22 introduced significant enhancements to `http.ServeMux`, enabling method-based routing and path wildcards (e.g., `GET /items/{id}`) natively.

| Feature | `chi` / `gin` | `net/http` (Go 1.22+) | Impact |
| :--- | :--- | :--- | :--- |
| **Routing** | `r.Get("/users/{id}", handler)` | `mux.HandleFunc("GET /users/{id}", handler)` | **High**: Zero-dep routing. |
| **Middleware** | `r.Use(middleware)` | Native `http.Handler` wrapping | **Medium**: Requires slightly more boilerplate for chains. |
| **Binary Size** | Adds ~1-2MB (Gin is heavier) | Included in Go runtime | **High**: Significant reduction. |
| **Performance** | Optimized, but adds overhead | Highly optimized, zero allocation overhead | **Positive**: Faster cold starts. |

### Migration Strategy
1.  **Phase 1: Adapter Layer**: Create a compatibility layer in `internal/adapters/http` that mimics the `chi` interface but uses `http.NewServeMux()` under the hood.
2.  **Phase 2: Middleware Adaptation**: Rewrite middleware to use the standard `func(next http.Handler) http.Handler` signature (most `chi` middleware already follows this, but `gin` does not).
3.  **Phase 3: Template Update**: Update `kthulu-cli` templates to scaffold new projects using `net/http` by default.

**Code Example (Migration):**

*Before (`chi`):*
```go
r := chi.NewRouter()
r.Get("/users/{id}", GetUser)
```

*After (`net/http`):*
```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", GetUser)
// Parsing params:
// id := mux.PathValue(r, "id")
```

## 2. Logging Migration (`log/slog`)

### Current State
*   **`zap`**: Deeply integrated via `core/logger.go`. Used for structured logging and performance.

### The `log/slog` Advantage
`log/slog` is the standard structured logging library introduced in Go 1.21. It provides high performance and a unified interface, allowing libraries to log without forcing a specific backend on the user.

| Feature | `zap` | `log/slog` | Impact |
| :--- | :--- | :--- | :--- |
| **Performance** | Best-in-class | Very close to Zap | **Negligible**: Both are excellent. |
| **API** | `logger.Info("msg", zap.String("k", "v"))` | `logger.Info("msg", "k", "v")` | **Positive**: Simpler API. |
| **Dependency** | Large dependency tree | Zero dependency | **High**: Reduces build graph. |

### Migration Strategy
1.  **Refactor `core/logger.go`**: The current `Logger` interface wraps `zap.SugaredLogger`. We can reimplement this interface to wrap `*slog.Logger` instead.
2.  **Replace `NewLogger`**: Update the factory functions to return a `slog.Logger` configured with a JSON handler (production) or Text handler (development).
3.  **Context**: `slog` has native support for adding attributes to context, simplifying `WithRequestContext`.

**Code Example (Migration):**

*Before (`zap`):*
```go
logger.Info("user created", zap.String("id", "123"))
```

*After (`slog`):*
```go
logger.Info("user created", "id", "123")
```

## 3. BDD & Testing (Gherkin + Testify)

### Current State
*   `testify` is used for assertions (`require.NoError`, `assert.Equal`). This is good and should be kept as it improves test readability significantly over `if err != nil`.

### Introducing BDD (Behavior Driven Development)
To achieve "Cucumber-like" planning and testing, we recommend **Godog**. It parses standard Gherkin (`.feature`) files and executes them using Go code.

**Recommended Stack:**
*   **Specs**: `.feature` files (Gherkin syntax).
*   **Runner**: `github.com/cucumber/godog`.
*   **Assertions**: `github.com/stretchr/testify/assert` (within the step definitions).

### Workflow
1.  **Plan**: Write `features/login.feature` using Gherkin.
    ```gherkin
    Feature: User Login
      Scenario: Successful login
        Given a user "alice" exists with password "secure"
        When I send a POST request to "/login"
        Then the response code should be 200
    ```
2.  **Generate**: Use `godog` to generate step skeletons.
3.  **Implement**: Fill in the steps using `testify` for assertions.

**Example Step Implementation:**
```go
func (ctx *TestContext) iSendAPOSTRequestTo(endpoint string) error {
    // ... make request ...
    return nil
}

func (ctx *TestContext) theResponseCodeShouldBe(code int) error {
    assert.Equal(ctx.t, code, ctx.lastResponse.Code)
    return nil
}
```

## 4. Other Standard Library Opportunities

*   **`pkg/errors`**: Replace with standard `errors` (using `%w` wrapping).
*   **`google/uuid`**: Can be kept (de-facto standard) or replaced with `rand` if only simple unique strings are needed (though UUIDs are usually preferred for DB keys).
*   **`gorm`**: Keep for now if heavily used, but moving to `sql/sql` or `pgx` directly is the ultimate "standard lib" move for DBs (though high effort).

## Conclusion
Moving to `net/http` and `log/slog` will significantly reduce the project's footprint and future-proof the codebase. The `core` abstraction layers already present in the project make this feasible without a "big bang" rewrite.
