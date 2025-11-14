# Overrides and Extensions

Kthulu uses tagged annotations to mark customization points.
This guide explains how to **extend** existing behavior or **override** it entirely.

## Override vs. Extend

| Action    | Annotation           | Description |
|-----------|----------------------|-------------|
| Extend    | `@kthulu:wrap`       | Inject extra logic while calling the original implementation. Safe for most use cases. |
| Override  | `@kthulu:shadow`     | Replace the original implementation. Dangerous because core logic is skipped. |

> **Safety tip:** Prefer `@kthulu:wrap` when possible. Use `@kthulu:shadow` only when a complete replacement is required and you fully understand the consequences.

## Example: Extending `AuthenticateUser`

1. Create a new file under your module.
2. Add the `@kthulu:wrap` annotation.
3. Accept a `next` function that calls the original implementation.
4. Register the wrapper with Fx using `fx.Decorate` or similar.

```go
// auth/custom_auth.go
// @kthulu:wrap
package auth

func AuthenticateUser(ctx context.Context, creds Credentials, next AuthenticateUserFunc) (*User, error) {
    // pre-hook
    user, err := next(ctx, creds)
    if err != nil {
        return nil, err
    }
    // post-hook
    return user, nil
}
```

## Example: Overriding the Product Repository

1. Implement the `repository.ProductRepository` interface.
2. Annotate the file with `@kthulu:shadow`.
3. Replace the default binding with `fx.Replace`.
4. Run contract tests (`make test-contracts`) to confirm your implementation satisfies the interface.

```go
// repository/my_product_repo.go
// @kthulu:shadow
package repository

var _ ProductRepository = (*MyProductRepository)(nil)

type MyProductRepository struct { /* ... */ }

// Implement all methods from ProductRepository...
```

To verify the override, run the product repository contract test:

```sh
go test -v -run TestProductRepositoryContract ./internal/contracts/
```

## Warnings

- `@kthulu:shadow` bypasses core checks. Misuse can break critical paths.
- Always run contract tests after overriding a component.

## Contract Tests

Contract tests in `backend/internal/contracts` ensure replacements remain compatible with declared interfaces. They act as a safety net when using `@kthulu:shadow` and provide confidence that extensions preserve expected behavior.

Run all contract tests:

```sh
make test-contracts
```
