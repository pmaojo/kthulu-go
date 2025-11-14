# Repository Documentation

## Testing Framework

**Target Framework:** Playwright

The repository uses Playwright as the primary E2E testing framework with the following configuration:
- Test directory: `e2e/tests/`
- Configuration: `e2e/playwright.config.ts`
- Test utilities: `e2e/utils/test-helpers.ts` and `e2e/utils/api-client.ts`
- Global setup/teardown: `e2e/global-setup.ts` and `e2e/global-teardown.ts`
- Cross-browser testing: Chromium, Firefox, WebKit, Mobile Chrome, Mobile Safari
- Test timeout: 30 seconds
- Base URL: http://localhost:8080 (configurable via BASE_URL env var)

## Test Structure

The E2E tests are organized by feature areas:
- `auth.spec.ts` - Authentication flows
- `contacts.spec.ts` - Contact management
- `invoices.spec.ts` - Invoice management
- `organizations.spec.ts` - Organization management
- `products.spec.ts` - Product management
- `integration.spec.ts` - Full application integration tests

## Running Tests

Use the provided script for comprehensive testing:
```bash
./scripts/test-e2e.sh
```

Or run Playwright directly:
```bash
cd e2e && npx playwright test
```

## Test Data

Tests use:
- In-memory SQLite database for isolation
- Test users: admin@test.com, user@test.com, manager@test.com
- Password: TestPassword123!
- Global setup creates test users and resets database state