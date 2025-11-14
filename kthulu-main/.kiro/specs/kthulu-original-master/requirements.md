# Requirements Document

## Introduction

This specification defines the requirements for building the complete `kthulu-original` master project - a reference monolith containing all essential backend enterprise modules (ERP-lite) plus Vite frontend, 100% dockerized, that will serve as the source for the `kthulu` scaffolder CLI.

The master project must be fully functional and "deconstructible" - meaning every component can be selectively copied or disabled by the CLI generator through proper tagging and modular architecture.

## Requirements

### Requirement 1: Complete Module Coverage

**User Story:** As a framework developer, I want the master project to contain all core and ERP-lite modules, so that the CLI can generate projects with any combination of these modules.

#### Acceptance Criteria

1. WHEN the project is built THEN it SHALL include all core modules: auth, user, access, notifier
2. WHEN the project is built THEN it SHALL include all ERP-lite modules: org, contacts, products, invoices, inventory, calendar
3. WHEN each module is implemented THEN it SHALL provide minimum CRUD functionality
4. WHEN modules are implemented THEN they SHALL follow the defined dependency chain (auth → user → org → contacts → products, etc.)

### Requirement 2: Hexagonal Architecture Implementation

**User Story:** As a developer using the generated code, I want a clean hexagonal architecture, so that I can easily understand, test, and extend the codebase.

#### Acceptance Criteria

1. WHEN the project structure is created THEN it SHALL follow the pattern: adapters → usecase → repository → infrastructure
2. WHEN any module is implemented THEN it SHALL separate concerns into these distinct layers
3. WHEN repositories are created THEN interfaces SHALL be defined in `internal/repository` and implementations in `internal/infrastructure/db`
4. WHEN use cases are implemented THEN they SHALL only depend on repository interfaces, never concrete implementations
5. WHEN adapters are created THEN they SHALL handle HTTP concerns and delegate business logic to use cases

### Requirement 3: Dependency Injection with Auto-Discovery

**User Story:** As a framework user, I want modules to be automatically discoverable and configurable, so that I can easily enable/disable features without manual wiring.

#### Acceptance Criteria

1. WHEN each module is created THEN it SHALL export an `fx.Module` in its `module.go` file
2. WHEN modules register routes THEN they SHALL use a `RegisterRoutes` function that can be auto-discovered
3. WHEN modules register repositories THEN they SHALL use a `RegisterRepo` function that can be auto-discovered
4. WHEN the main application starts THEN it SHALL automatically wire all enabled modules through fx dependency injection
5. WHEN a module is disabled THEN the application SHALL start successfully without it

### Requirement 4: Frontend Integration

**User Story:** As an end user, I want a complete frontend application connected to the API, so that I can interact with all the backend functionality through a web interface.

#### Acceptance Criteria

1. WHEN the frontend is built THEN it SHALL use Vite + React + TypeScript + Tailwind CSS
2. WHEN authentication is implemented THEN it SHALL provide `/login`, `/register`, and `/profile` pages
3. WHEN the frontend is configured THEN it SHALL include Zustand for state management and TanStack Query for data fetching
4. WHEN API calls are made THEN they SHALL use a centralized axios client with automatic token refresh
5. WHEN the frontend starts THEN it SHALL successfully connect to the backend API

### Requirement 5: Docker Compose Development Environment

**User Story:** As a developer, I want to start the entire stack with one command, so that I can quickly set up a development environment.

#### Acceptance Criteria

1. WHEN `docker-compose up` is executed THEN it SHALL start API on port 8080, Web on port 5173, and Postgres on port 5432
2. WHEN containers start THEN the API SHALL wait for the database to be ready before starting
3. WHEN the database starts THEN migrations SHALL run automatically
4. WHEN all services are running THEN the frontend SHALL be able to communicate with the API
5. WHEN services are stopped and restarted THEN data SHALL persist in the database

### Requirement 6: OpenAPI Specification Generation

**User Story:** As an API consumer, I want automatically generated API documentation, so that I can understand and integrate with all available endpoints.

#### Acceptance Criteria

1. WHEN the API is built THEN it SHALL automatically generate an OpenAPI 3.1 specification
2. WHEN the specification is generated THEN it SHALL include all endpoints from all modules
3. WHEN DTOs are defined THEN they SHALL be automatically included in the OpenAPI schema
4. WHEN the specification is updated THEN it SHALL be committed to version control
5. WHEN the API is running THEN Swagger UI SHALL be available at `/docs`

### Requirement 7: Type Safety Between Go and TypeScript

**User Story:** As a full-stack developer, I want type safety between backend and frontend, so that I can catch integration errors at compile time.

#### Acceptance Criteria

1. WHEN Go DTOs are defined THEN TypeScript interfaces SHALL be automatically generated
2. WHEN the generation script runs THEN it SHALL produce `frontend/src/types/kthulu-api.ts`
3. WHEN TypeScript types are generated THEN Zod schemas SHALL also be created for runtime validation
4. WHEN API contracts change THEN the build SHALL fail if types are not regenerated
5. WHEN frontend code uses API types THEN it SHALL have full IntelliSense support

### Requirement 8: Contract Testing

**User Story:** As a developer, I want automated tests that verify interface compliance, so that I can catch implementation drift early.

#### Acceptance Criteria

1. WHEN repository interfaces are defined THEN contract tests SHALL verify implementations satisfy the interface
2. WHEN contract tests run THEN they SHALL use httptest for HTTP layer testing
3. WHEN any repository implementation changes THEN contract tests SHALL catch breaking changes
4. WHEN tests are executed THEN they SHALL provide clear feedback on which contracts are violated
5. WHEN the test suite runs THEN contract tests SHALL be included in the overall test coverage

### Requirement 9: CLI Deconstructibility Tagging

**User Story:** As a CLI tool developer, I want to identify which files belong to the core framework, so that I can selectively copy them when generating new projects.

#### Acceptance Criteria

1. WHEN core framework files are created THEN they SHALL include the tag `// @kthulu:core`
2. WHEN module-specific files are created THEN they SHALL include appropriate module tags
3. WHEN the CLI processes files THEN it SHALL be able to filter by these tags
4. WHEN files are tagged THEN the tagging SHALL be consistent across all file types (Go, TypeScript, SQL, etc.)
5. WHEN tags are applied THEN they SHALL not interfere with normal compilation or execution

### Requirement 10: Comprehensive Documentation

**User Story:** As a new developer on the project, I want clear documentation on how to set up, run, and extend the system, so that I can be productive quickly.

#### Acceptance Criteria

1. WHEN the README is created THEN it SHALL include setup instructions for development with Docker Compose
2. WHEN documentation is written THEN it SHALL explain how to run `make dev`, `make test`, and `make gen-types`
3. WHEN the system is documented THEN it SHALL include an end-to-end demo scenario
4. WHEN new features are added THEN documentation SHALL be updated accordingly
5. WHEN developers read the documentation THEN they SHALL be able to successfully set up and run the project

### Requirement 11: Spanish Tax Compliance (Veri*Factu)

**User Story:** As a Spanish business owner, I want my invoicing system to comply with Veri*Factu regulations (RD 1007/2023 RRSIF), so that I can meet legal requirements for invoice verification and AEAT submission.

#### Acceptance Criteria

1. WHEN an invoice is created THEN the system SHALL generate a structured record (XML/JSON) according to RRSIF specifications
2. WHEN Veri*Factu mode is enabled THEN the system SHALL submit invoice records to AEAT in real-time
3. WHEN non-Veri*Factu mode is enabled THEN the system SHALL create digital signatures and queue records for later submission
4. WHEN invoices are generated THEN they SHALL include QR codes with "VERI*FACTU" verification and legal legends
5. WHEN invoice operations occur THEN the system SHALL maintain an audit trail of all events (alta, baja, incidents)
6. WHEN network failures occur THEN the system SHALL implement automatic retry mechanisms with incident logging
7. WHEN the module is disabled THEN the system SHALL continue normal invoice operations without compliance features
8. WHEN multiple terminals are used THEN each SHALL operate independently while maintaining separate compliance records

### Requirement 12: Enhanced Storage Abstraction

**User Story:** As a system architect, I want flexible storage interfaces for tokens and caching, so that I can easily switch between different storage backends (memory, Redis, database) without changing business logic.

#### Acceptance Criteria

1. WHEN storage interfaces are defined THEN they SHALL provide generic operations (get, set, delete, exists)
2. WHEN token storage is implemented THEN it SHALL support token revocation and expiry management
3. WHEN cache storage is implemented THEN it SHALL provide increment/decrement and TTL operations
4. WHEN storage backends are changed THEN business logic SHALL remain unchanged
5. WHEN token providers are used THEN they SHALL integrate seamlessly with HTTP clients

### Requirement 13: Comprehensive Pagination System

**User Story:** As an API consumer, I want consistent pagination across all endpoints, so that I can efficiently handle large datasets with search and filtering capabilities.

#### Acceptance Criteria

1. WHEN pagination is implemented THEN it SHALL provide consistent parameters (page, pageSize, sortBy, sortDir)
2. WHEN paginated results are returned THEN they SHALL include metadata (total, totalPages, hasNext, hasPrev)
3. WHEN search is performed THEN it SHALL work seamlessly with pagination
4. WHEN filtering is applied THEN it SHALL integrate with pagination and search
5. WHEN repositories implement pagination THEN they SHALL use the generic PaginationResult type

### Requirement 14: Complete ERP-lite Module Suite

**User Story:** As a business user, I want complete inventory and calendar management capabilities, so that I can run a comprehensive ERP system.

#### Acceptance Criteria

1. WHEN inventory module is implemented THEN it SHALL support multi-warehouse stock management
2. WHEN stock movements occur THEN the system SHALL maintain complete audit trails
3. WHEN calendar module is implemented THEN it SHALL support appointment scheduling with conflict detection
4. WHEN appointments are booked THEN the system SHALL send notifications and reminders
5. WHEN modules are integrated THEN they SHALL work seamlessly with existing ERP-lite modules

### Requirement 15: Enterprise-Grade Testing Framework

**User Story:** As a quality assurance engineer, I want comprehensive testing coverage, so that I can ensure system reliability and catch regressions early.

#### Acceptance Criteria

1. WHEN contract tests are implemented THEN they SHALL verify all repository interfaces
2. WHEN integration tests run THEN they SHALL test real database interactions
3. WHEN API tests execute THEN they SHALL validate all HTTP endpoints
4. WHEN tests are run THEN they SHALL provide coverage reports and metrics
5. WHEN CI/CD pipeline runs THEN all tests SHALL pass before deployment

### Requirement 16: Complete Frontend Application

**User Story:** As an end user, I want a fully functional web application, so that I can manage all business operations through an intuitive interface.

#### Acceptance Criteria

1. WHEN frontend loads THEN it SHALL provide responsive design across all devices
2. WHEN users interact THEN the interface SHALL provide real-time feedback and loading states
3. WHEN data is fetched THEN it SHALL use optimized caching and background updates
4. WHEN forms are submitted THEN they SHALL provide validation and error handling
5. WHEN accessibility is tested THEN it SHALL meet WCAG 2.1 AA standards

### Requirement 17: Advanced Security and Authentication

**User Story:** As a security administrator, I want enterprise-grade authentication features, so that I can ensure system security and compliance.

#### Acceptance Criteria

1. WHEN OAuth2 is configured THEN users SHALL be able to login with external providers
2. WHEN 2FA is enabled THEN users SHALL complete multi-factor authentication
3. WHEN security events occur THEN they SHALL be logged and monitored
4. WHEN password policies are set THEN they SHALL be enforced consistently
5. WHEN sessions are managed THEN they SHALL support device tracking and revocation

### Requirement 18: Production-Ready Observability

**User Story:** As a system administrator, I want comprehensive monitoring and observability, so that I can maintain system health and performance.

#### Acceptance Criteria

1. WHEN metrics are collected THEN they SHALL be exposed in Prometheus format
2. WHEN traces are generated THEN they SHALL provide end-to-end request tracking
3. WHEN errors occur THEN they SHALL be tracked and alerted appropriately
4. WHEN performance degrades THEN monitoring SHALL detect and alert on issues
5. WHEN dashboards are viewed THEN they SHALL provide actionable insights