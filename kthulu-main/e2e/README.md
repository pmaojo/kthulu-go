# Kthulu E2E Testing Framework

This directory contains comprehensive end-to-end (E2E) tests for the Kthulu application using Playwright.

## Overview

The E2E testing framework provides:

- **Full-stack testing** of the complete application workflow
- **Cross-browser testing** (Chrome, Firefox, Safari, Mobile)
- **API integration testing** with the backend
- **Visual regression testing** with screenshots
- **Performance testing** and monitoring
- **Test data management** and cleanup
- **CI/CD integration** with detailed reporting

## Test Structure

```
e2e/
├── tests/                  # Test files
│   ├── auth.spec.ts       # Authentication flow tests
│   ├── organizations.spec.ts # Organization management tests
│   ├── contacts.spec.ts   # Contact management tests
│   ├── products.spec.ts   # Product management tests
│   ├── invoices.spec.ts   # Invoice management tests
│   └── integration.spec.ts # Full application integration tests
├── utils/                 # Test utilities
│   ├── api-client.ts     # API client for backend integration
│   └── test-helpers.ts   # Common test helper functions
├── global-setup.ts       # Global test setup
├── global-teardown.ts    # Global test cleanup
├── playwright.config.ts  # Playwright configuration
└── package.json          # Dependencies and scripts
```

## Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn
- Backend server running on `http://localhost:8080`

### Installation

```bash
# Install dependencies
npm install

# Install Playwright browsers
npx playwright install

# Install system dependencies (Linux only)
npx playwright install-deps
```

### Running Tests

```bash
# Run all tests
npm test

# Run tests in headed mode (see browser)
npm run test:headed

# Run tests with UI mode
npm run test:ui

# Run specific test file
npx playwright test auth.spec.ts

# Run tests in debug mode
npm run test:debug

# Generate test code
npm run test:codegen
```

### Scaffolding New Tests

Generate a new test spec and corresponding Page Object with:

```bash
npm run test:scaffold
```

The command prompts for a module name and creates:

- `tests/<module>.spec.ts` – base test file with setup hooks
- `pages/<module>-page.ts` – matching Page Object class

Use these files as starting points and expand them with real test cases and helper methods.

## Test Categories

### 1. Authentication Tests (`auth.spec.ts`)

Tests the complete authentication flow:

- User registration with validation
- Login with valid/invalid credentials
- Password reset functionality
- Session persistence and token management
- Logout and session cleanup
- Concurrent session handling

### 2. Organization Management (`organizations.spec.ts`)

Tests organization CRUD operations:

- Creating new organizations
- Editing organization details
- Deleting organizations
- Organization search and filtering
- Pagination handling
- Permission-based access control

### 3. Contact Management (`contacts.spec.ts`)

Tests contact management features:

- Creating customers, suppliers, leads, partners
- Contact information management
- Address and phone number management
- Contact search and filtering
- Contact status management
- Data export functionality

### 4. Product Management (`products.spec.ts`)

Tests product catalog features:

- Product creation and editing
- SKU management and validation
- Product variants and pricing
- Category filtering
- Bulk operations
- Import/export functionality

### 5. Invoice Management (`invoices.spec.ts`)

Tests invoicing functionality:

- Invoice creation and editing
- Invoice item management
- Payment recording
- Status management
- PDF generation
- Email sending
- Statistics and reporting

### 6. Integration Tests (`integration.spec.ts`)

Tests complete business workflows:

- End-to-end business processes
- Multi-organization workflows
- Concurrent user sessions
- Error handling and recovery
- Offline scenarios
- Performance under load
- Data consistency across modules

## Test Utilities

### API Client (`utils/api-client.ts`)

Provides a comprehensive API client for backend integration:

```typescript
const apiClient = new ApiClient('http://localhost:8080');

// Authentication
await apiClient.register('user@test.com', 'password');
await apiClient.login('user@test.com', 'password');

// Organizations
const org = await apiClient.createOrganization('Test Org');

// Contacts
const contact = await apiClient.createContact(org.id, {
  companyName: 'Test Contact',
  email: 'contact@test.com',
  type: 'customer'
});
```

### Test Helpers (`utils/test-helpers.ts`)

Provides common test helper functions:

```typescript
const helpers = new TestHelpers(page, apiClient);

// Authentication helpers
await helpers.loginAsAdmin();
await helpers.loginAsUser();

// Navigation helpers
await helpers.navigateToContacts();
await helpers.navigateToProducts();

// Form helpers
await helpers.fillForm({
  name: 'Test Name',
  email: 'test@example.com'
});

// Assertion helpers
await helpers.expectToBeOnPage('/dashboard');
await helpers.expectElementToBeVisible('data-table');
```

## Configuration

### Playwright Configuration (`playwright.config.ts`)

Key configuration options:

```typescript
export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
    { name: 'webkit', use: { ...devices['Desktop Safari'] } },
    { name: 'Mobile Chrome', use: { ...devices['Pixel 5'] } },
    { name: 'Mobile Safari', use: { ...devices['iPhone 12'] } },
  ],
});
```

### Environment Variables

- `BASE_URL` - Base URL of the application (default: `http://localhost:8080`)
- `CI` - Set to `true` in CI environments for optimized settings
- `HEADLESS` - Run tests in headless mode (default: `true`)

## Test Data Management

### Global Setup

The global setup (`global-setup.ts`) handles:

- Server health checks
- Database reset and seeding
- Test user creation
- Initial data setup

### Test Isolation

Each test:

- Runs in isolation with fresh data
- Uses unique test data to avoid conflicts
- Cleans up after execution
- Handles concurrent test execution

### Data Factories

Helper functions create consistent test data:

```typescript
// Create test organization
const orgId = await helpers.createTestOrganization('Test Org');

// Create test contact
const contactId = await helpers.createTestContact(orgId, 'Test Contact');

// Create test product
const productId = await helpers.createTestProduct(orgId, 'Test Product');
```

## Reporting

### HTML Report

Playwright generates comprehensive HTML reports:

```bash
# View last test report
npm run test:report
```

### CI/CD Integration

The framework supports multiple report formats:

- **HTML** - Interactive report with screenshots and videos
- **JSON** - Machine-readable results for CI/CD
- **JUnit** - XML format for test result integration

### Screenshots and Videos

- Screenshots captured on test failure
- Videos recorded for failed tests
- Traces available for debugging

## Best Practices

### Test Organization

1. **Group related tests** in describe blocks
2. **Use descriptive test names** that explain the scenario
3. **Keep tests independent** and isolated
4. **Use page object patterns** for complex interactions

### Data Management

1. **Create fresh test data** for each test
2. **Use unique identifiers** to avoid conflicts
3. **Clean up after tests** to prevent side effects
4. **Use realistic test data** that matches production scenarios

### Assertions

1. **Use specific assertions** rather than generic ones
2. **Wait for elements** before interacting
3. **Assert on meaningful state** not just presence
4. **Provide clear error messages** for failed assertions

### Performance

1. **Run tests in parallel** when possible
2. **Use efficient selectors** (data-testid preferred)
3. **Minimize network requests** in setup
4. **Reuse authentication** across tests

## Debugging

### Debug Mode

```bash
# Run in debug mode
npm run test:debug

# Debug specific test
npx playwright test auth.spec.ts --debug
```

### UI Mode

```bash
# Run with UI
npm run test:ui
```

### Trace Viewer

```bash
# View traces for failed tests
npx playwright show-trace test-results/trace.zip
```

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Run E2E tests
  run: |
    npm install
    npx playwright install --with-deps
    npm run test
  env:
    CI: true
    BASE_URL: http://localhost:8080
```

### Docker Integration

```dockerfile
FROM mcr.microsoft.com/playwright:v1.40.0-focal

COPY e2e/ /app/e2e/
WORKDIR /app/e2e

RUN npm install
CMD ["npm", "test"]
```

## Troubleshooting

### Common Issues

1. **Server not ready** - Increase timeout in global setup
2. **Flaky tests** - Add proper waits and retries
3. **Element not found** - Check selectors and timing
4. **Authentication issues** - Verify test user setup

### Debug Tips

1. **Use headed mode** to see what's happening
2. **Add screenshots** at key points
3. **Check network tab** for API errors
4. **Verify test data** is created correctly

## Contributing

When adding new tests:

1. Follow the existing test structure
2. Use the provided utilities and helpers
3. Add appropriate assertions and error handling
4. Update documentation for new test scenarios
5. Ensure tests are reliable and maintainable

## Performance Benchmarks

The E2E tests include performance monitoring:

- Page load times
- API response times
- User interaction responsiveness
- Memory usage tracking

Results are included in test reports for performance regression detection.