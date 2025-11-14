# Implementation Plan

> **Current State:** The project already has a solid foundation with basic infrastructure, Docker setup, and frontend scaffolding in place. This plan focuses on completing the missing pieces and enhancing existing components to meet all requirements.

**Legend:**

- ✅ = Already implemented/exists
- [ ] = Needs to be implemented
- [x] = Completed/exists but may need minor enhancements

- [x] 1. Project Foundation Setup

  - ✅ Complete directory structure already exists
  - ✅ Go module initialized with proper dependencies (Chi, Fx, Zap, GORM, etc.)
  - ✅ Basic configuration files in place (.gitignore, docker-compose.yml, Makefile)
  - _Requirements: 1.1, 2.1, 10.1_

- [ ] 2. Core Infrastructure Enhancement

  - [x] 2.1 Enhance core configuration system

    - ✅ Basic config.go exists, extend with server, SMTP, and additional JWT configuration
    - Add `// @kthulu:core` tags to existing core files for CLI deconstructibility
    - Add validation and better error handling for configuration
    - _Requirements: 2.2, 9.1_

  - [x] 2.2 Enhance logging system

    - ✅ Basic logger.go exists, extend with middleware integration
    - Add structured logging utilities and request correlation IDs
    - Add `// @kthulu:core` tags and improve development/production configuration
    - _Requirements: 2.2, 9.1_

  - [x] 2.3 Complete database connection and migration system
    - ✅ Basic db.go and migrate.go exist, enhance with connection pooling
    - Implement comprehensive health checks and connection retry logic
    - Add `// @kthulu:core` tags and improve error handling
    - _Requirements: 2.2, 5.3, 9.1_

- [-] 3. Authentication and User Management Core

  - [x] 3.1 Create user domain entities and repository interfaces

    - Define User, Role, and RefreshToken entities in domain layer
    - Create repository interfaces in `internal/repository/`
    - Implement GORM repository implementations in `internal/infrastructure/db/`
    - _Requirements: 1.1, 2.3, 2.4_

  - [x] 3.2 Implement authentication use cases

    - Create register, login, confirm, refresh, and logout use cases
    - Implement JWT token generation and validation
    - Add password hashing and email confirmation logic
    - _Requirements: 1.1, 2.4, 4.2_

  - [x] 3.3 Create authentication HTTP handlers
    - Implement REST endpoints for auth module
    - Add request/response DTOs with validation
    - Create middleware for JWT token verification
    - _Requirements: 1.1, 2.1, 4.2_

- [x] 4. User Profile and Access Control

  - [x] 4.1 Implement user profile management

    - Create user profile use cases (get profile, update profile)
    - Implement HTTP handlers for `/users/me` endpoints
    - Add profile update validation and business rules
    - _Requirements: 1.1, 4.2_

  - [x] 4.2 Implement role-based access control

    - Create roles and permissions domain entities
    - Implement RBAC middleware with X-Role-Scope header support
    - Add role assignment and permission checking logic
    - _Requirements: 1.1, 2.4_

  - [x] 4.3 Create notification system stub
    - Implement notification interface and console provider
    - Create email notification templates and sending logic
    - Integrate notification system with authentication flow
    - _Requirements: 1.1, 4.4_

- [x] 5. Fx Module System and Auto-Discovery

  - [x] 5.1 Complete module registration system

    - ✅ Basic Fx setup exists in main.go, complete module.go files for each core module
    - Implement auto-discovery mechanism for routes and repositories
    - Complete Fx dependency injection wiring with all modules
    - _Requirements: 3.1, 3.2, 3.4_

  - [x] 5.2 Enhance main application entry point
    - ✅ Basic main.go with Fx exists, complete with all module registrations
    - Add health check endpoints and proper graceful shutdown
    - Implement comprehensive error handling and logging
    - _Requirements: 3.4, 6.5_

- [-] 6. ERP-lite Modules Implementation

  - [x] 6.1 Implement organization module

    - Create organization entities (Organization, OrganizationUser, Invitation)
    - Implement multi-tenant organization management use cases
    - Add HTTP handlers for organization CRUD operations
    - _Requirements: 1.1, 1.4, 2.1_

  - [x] 6.2 Implement contacts module

    - ✅ Create contact entities (Contact, ContactAddress, ContactPhone)
    - ✅ Implement customer/supplier management use cases
    - ✅ Add HTTP handlers for contact management with tenant isolation
    - ✅ Complete CRUD operations with filtering, pagination, and search
    - ✅ Integrated with Fx dependency injection and route registry
    - _Requirements: 1.1, 1.4, 2.1_

  - [x] 6.3 Implement products module

    - Create product entities (Product, ProductVariant, ProductPrice)
    - Implement product catalog management use cases
    - Add HTTP handlers for product CRUD with pricing and variants
    - _Requirements: 1.1, 1.4, 2.1_

  - [x] 6.4 Implement invoices module

    - ✅ Create invoice entities (Invoice, InvoiceItem, Payment)
    - ✅ Implement invoice generation and payment tracking use cases
    - ✅ Add HTTP handlers for invoice management with comprehensive CRUD
    - ✅ Implement advanced filtering, pagination, and statistics
    - ✅ Complete integration with Fx dependency injection
    - _Requirements: 1.1, 1.4, 2.1_

  - [x] 6.5 Implement inventory module

    - Create inventory entities (Warehouse, InventoryItem, StockMovement)
    - Implement stock management and movement tracking use cases
    - Add HTTP handlers for inventory operations
    - _Requirements: 1.1, 1.4, 2.1_

  - [x] 6.6 Implement calendar module

    - Create calendar entities (Event, Appointment, AvailabilitySlot)
    - Implement appointment scheduling and availability use cases
    - Add HTTP handlers for calendar management
      https://github.com/rickar/cal
    - _Requirements: 1.1, 1.4, 2.1_

  - [ ] 6.7 Implement Veri\*Factu compliance module (Spanish tax compliance)
    - Create Veri\*Factu domain entities (VeriFactuRecord, VeriFactuEvent, VeriFactuSignature)
    - Implement AEAT real-time submission and queued mode with digital signatures
    - Add QR code generation and legal legend integration for invoices
    - Create XML/JSON structured record generation according to RD 1007/2023 RRSIF
    - Implement audit trail and event logging for compliance
    - Add HTTP handlers for Veri\*Factu management and webhook endpoints
    - _Requirements: 1.1, 1.4, 2.1, Spanish Tax Compliance_

- [x] 7. Database Migrations and Schema

  - Create Goose migration files for all entities
  - Implement proper foreign key relationships and constraints
  - Add database indexes for performance optimization
  - Create seed data for development and testing
  - _Requirements: 5.3, 1.4_

- [x] 8. OpenAPI Specification Generation

  - [x] 8.1 Create OpenAPI generation script

    - ✅ Implement `scripts/openapi.sh` using swag (swagger generation)
    - ✅ Configure OpenAPI spec generation from Go structs and handlers
    - ✅ Add module tagging for selective API generation
    - ✅ Fixed script syntax errors and made it executable
    - _Requirements: 6.1, 6.2, 6.4_

  - [x] 8.2 Generate complete API specification
    - ✅ Run OpenAPI generation for all modules
    - ✅ Generated specification includes auth, user, organization, and health endpoints
    - ✅ Committed generated `api/openapi.yaml` to version control
    - _Requirements: 6.2, 6.4_

- [x] 9. TypeScript Type Generation

  - [x] 9.1 Create TypeScript generation script

    - ✅ Implement `scripts/gen-types.sh` with fallback approach
    - ✅ Created manual TypeScript interfaces based on OpenAPI spec
    - ✅ Add Zod schema generation for runtime validation
    - ✅ Installed zod dependency in frontend
    - _Requirements: 7.1, 7.2, 7.3_

  - [x] 9.2 Generate and integrate TypeScript types
    - ✅ Created `frontend/src/types/kthulu-api.ts` with comprehensive types
    - ✅ Created `frontend/src/types/kthulu-api-zod.ts` with validation schemas
    - ✅ Created `frontend/src/types/index.ts` for easy imports
    - ✅ Types match backend DTOs and include all auth, user, and organization models
    - _Requirements: 7.2, 7.5_

- [ ] 10. Contract Testing Implementation

  - [x] 10.1 Create contract testing framework

    - Set up contract test structure in `internal/contracts/`
    - Implement interface compliance tests for all repositories
    - Add httptest integration for HTTP layer contract testing
    - _Requirements: 8.1, 8.2, 8.4_

  - [x] 10.2 Implement comprehensive contract tests
    - Write contract tests for all repository interfaces
    - Create HTTP contract tests for all API endpoints
    - Add test coverage reporting and validation
    - _Requirements: 8.2, 8.3, 8.5_

- [ ] 11. Frontend Application Development

  - [x] 11.1 Set up Vite React TypeScript project

    - ✅ Frontend project initialized with Vite, React 18, and TypeScript
    - ✅ ESLint, Prettier, and Husky configured with pre-commit hooks
    - ✅ Project structure with proper folder organization exists
    - _Requirements: 4.1, 4.2_

  - [x] 11.2 Implement authentication pages and flow

    - ✅ Updated auth service with comprehensive authentication methods
    - ✅ Enhanced useAuth hook with full auth flow (login, register, confirm, refresh, logout)
    - ✅ Updated token storage to support access and refresh tokens
    - ✅ Added proper error handling and loading states
    - [ ] Create login, register, and profile pages with proper styling (UI components pending)
    - _Requirements: 4.2, 4.4, 4.5_

  - [x] 11.3 Integrate TanStack Query for API communication

    - Set up TanStack Query client configuration
    - Create API service functions using generated TypeScript types
    - Implement data fetching hooks for all authentication endpoints
    - _Requirements: 4.2, 4.4, 7.5_

  - [x] 11.4 Create responsive UI components
    - Implement shared UI components (buttons, forms, modals)
    - Create layout components (AdminLayout, PublicLayout)
    - Add proper error handling and loading states
    - _Requirements: 4.1, 4.2_

## 12. Enhanced Architecture Features

### 12.1 Storage Interface and Token Management

- **Status**: ✅ Completed
- **Description**:
  - ✅ Created generic storage interfaces (Storage, TokenStorage, CacheStorage)
  - ✅ Implemented memory-based token storage with automatic cleanup
  - ✅ Added token provider pattern for HTTP client integration
  - ✅ Enhanced auth service with token revocation and validation
  - ✅ Integrated storage abstraction into dependency injection system
- **Requirements**: Enhanced architecture
- **Dependencies**: 3.2 (Authentication)
- **Files created**:
  - `internal/repository/storage.go`
  - `internal/infrastructure/storage/memory_token_storage.go`
  - `internal/usecase/auth_service.go`

### 12.2 Pagination System

- **Status**: ✅ Completed
- **Description**:
  - ✅ Created comprehensive pagination infrastructure
  - ✅ Added pagination helper with query building, search, and filtering
  - ✅ Updated all repository interfaces with paginated methods
  - ✅ Implemented pagination in user, product, and invoice repositories
  - ✅ Added generic PaginationResult type with metadata
- **Requirements**: Enhanced architecture
- **Dependencies**: All repository implementations
- **Files created**:
  - `internal/infrastructure/db/pagination_helper.go`
  - Updated repository interfaces and implementations

## 13. Optional Compliance Modules

### 13.1 Implement Veri\*Factu module (Spain)

- **Status**: 📋 Specified and Ready for Implementation
- **Description**:
  - ✅ Complete specification created in `verifactu-extension.md`
  - ✅ Functional requirements (RF-VF-01 to RF-VF-05) defined
  - ✅ Architectural design (D-VF-01 to D-VF-04) documented
  - ✅ Implementation tasks (T-VF-01 to T-VF-07) planned
  - ✅ Integration points with invoice module identified
  - 📋 Ready for development phase
- **Requirements**: 2.7, Spanish Tax Compliance
- **Dependencies**: 6.4 (Invoices module)
- **Estimated effort**: 5-9 weeks
- **Specification**: `.kiro/specs/kthulu-original-master/verifactu-extension.md`

## 14. Complete ERP-lite Module Suite

### 14.1 Implement inventory module

- **Status**: ❌ Not Started
- **Description**:
  - Create inventory domain entities (Warehouse, InventoryItem, StockMovement, StockAdjustment)
  - Implement comprehensive stock management use cases (receive, transfer, adjust, reserve)
  - Add multi-warehouse support with location tracking
  - Create HTTP handlers for inventory operations with advanced filtering
  - Implement stock movement tracking and audit trail
  - Add low-stock alerts and reorder point management
  - Create inventory valuation methods (FIFO, LIFO, Average Cost)
  - Integrate with product module for stock tracking
- **Requirements**: 1.1, 1.4, 2.1
- **Dependencies**: 6.3 (Products module)
- **Estimated effort**: 3-4 weeks
- **Files to create**:
  - `internal/domain/inventory.go`
  - `internal/usecase/inventory.go`
  - `internal/adapters/http/inventory_handler.go`
  - `internal/modules/inventory.go`
  - `migrations/0008_create_inventory_extended.sql`

### 14.2 Implement calendar module

- **Status**: ❌ Not Started
- **Description**:
  \_ CLONE THIS REPO AND REUTILIZE FILES:
  https://github.com/rickar/cal
  - Create calendar domain entities (Event, Appointment, AvailabilitySlot, Calendar, Booking)
  - Implement appointment scheduling with conflict detection
  - Add recurring events and availability patterns
  - Create HTTP handlers for calendar management with timezone support
  - Implement availability checking and slot booking
  - Add calendar sharing and permission management
  - Create notification integration for appointment reminders
  - Support multiple calendar types (personal, shared, resource)
- **Requirements**: 1.1, 1.4, 2.1
- **Dependencies**: 4.1 (User module), 6.1 (Organization module)
- **Estimated effort**: 3-4 weeks
- **Files to create**:
  - `internal/domain/calendar.go`
  - `internal/usecase/calendar.go`
  - `internal/adapters/http/calendar_handler.go`
  - `internal/modules/calendar.go`
  - `migrations/0009_create_calendar.sql`

## 15. Contract Testing Framework

### 15.1 Create comprehensive contract testing framework

- **Status**: ❌ Not Started
- **Description**:
  - Set up contract test structure in `internal/contracts/`
  - Create interface compliance tests for all repository interfaces
  - Implement HTTP contract tests using httptest with golden files
  - Add behavioral contract tests for use cases
  - Create test utilities for database setup and teardown
  - Implement contract test runner with coverage reporting
  - Add contract validation for OpenAPI specifications
  - Create mock generators for testing isolation
- **Requirements**: 8.1, 8.2, 8.4
- **Dependencies**: All repository and use case implementations
- **Estimated effort**: 2-3 weeks
- **Files to create**:
  - `internal/contracts/repository_contracts_test.go`
  - `internal/contracts/http_contracts_test.go`
  - `internal/contracts/usecase_contracts_test.go`
  - `internal/testutils/`

### 15.2 Implement integration testing suite

- **Status**: ❌ Not Started
- **Description**:
  - Create end-to-end integration tests using testcontainers
  - Implement API integration tests for all modules
  - Add database integration tests with real PostgreSQL
  - Create authentication flow integration tests
  - Implement multi-module interaction tests
  - Add performance benchmarks for critical paths
  - Create test data factories and fixtures
  - Add CI/CD integration test pipeline
- **Requirements**: 8.2, 8.3, 8.5
- **Dependencies**: 15.1 (Contract testing framework)
- **Estimated effort**: 2-3 weeks
- **Files to create**:
  - `internal/integration/`
  - `internal/testdata/`
  - `scripts/test-integration.sh`

## 16. Frontend Application Completion

### 16.1 Complete TanStack Query integration

- **Status**: ❌ Not Started
- **Description**:
  - Set up TanStack Query client with proper configuration
  - Create API service functions using generated TypeScript types
  - Implement data fetching hooks for all modules (auth, user, org, contact, product, invoice)
  - Add optimistic updates and cache invalidation strategies
  - Create error handling and retry mechanisms
  - Implement infinite queries for paginated data
  - Add offline support and background sync
  - Create query devtools integration for development
- **Requirements**: 4.2, 4.4, 7.5
- **Dependencies**: 9.2 (TypeScript types)
- **Estimated effort**: 2-3 weeks
- **Files to create**:
  - `frontend/src/api/`
  - `frontend/src/hooks/queries/`
  - `frontend/src/utils/queryClient.ts`

### 16.2 Create comprehensive UI component library

- **Status**: ❌ Not Started
- **Description**:
  - Implement design system with Tailwind CSS components
  - Create reusable UI components (Button, Input, Modal, Table, Form, etc.)
  - Add accessibility support (ARIA labels, keyboard navigation)
  - Create layout components (AdminLayout, PublicLayout, Sidebar, Header)
  - Implement responsive design patterns
  - Add dark mode support with theme switching
  - Create component documentation with Storybook
  - Add component testing with React Testing Library
- **Requirements**: 4.1, 4.2
- **Dependencies**: None
- **Estimated effort**: 3-4 weeks
- **Files to create**:
  - `frontend/src/components/ui/`
  - `frontend/src/components/layouts/`
  - `frontend/src/styles/components.css`
  - `frontend/.storybook/`

### 16.3 Implement complete application pages

- **Status**: ❌ Not Started
- **Description**:
  - Create styled authentication pages (Login, Register, Confirm, ForgotPassword)
  - Implement user profile and settings pages
  - Create organization management interface
  - Add contact management with advanced filtering and search
  - Implement product catalog with variants and pricing
  - Create invoice management with PDF generation
  - Add inventory management interface (if inventory module completed)
  - Create calendar interface with appointment booking
  - Add dashboard with analytics and KPIs
- **Requirements**: 4.2, 4.4, 4.5
- **Dependencies**: 16.1 (TanStack Query), 16.2 (UI components)
- **Estimated effort**: 4-5 weeks
- **Files to create**:
  - `frontend/src/pages/`
  - `frontend/src/features/`
  - `frontend/src/components/views/`

## 17. Advanced Features and Polish

### 17.1 Implement advanced authentication features

- **Status**: ❌ Not Started
- **Description**:
  - Add OAuth2 integration (Google, GitHub, Microsoft)
  - Implement two-factor authentication (TOTP, SMS)
  - Add session management with device tracking
  - Create password policy enforcement
  - Implement account lockout and rate limiting
  - Add audit logging for security events
  - Create password reset with secure tokens
  - Add email verification with customizable templates
- **Requirements**: Enhanced security
- **Dependencies**: 3.2 (Authentication core)
- **Estimated effort**: 2-3 weeks
- **Files to create**:
  - `internal/usecase/oauth.go`
  - `internal/usecase/mfa.go`
  - `internal/adapters/http/oauth_handler.go`

### 17.2 Add observability and monitoring

- **Status**: ❌ Not Started
- **Description**:
  - Implement comprehensive metrics with Prometheus
  - Add distributed tracing with OpenTelemetry
  - Create health checks for all dependencies
  - Add structured logging with correlation IDs
  - Implement error tracking and alerting
  - Create performance monitoring and profiling
  - Add database query monitoring
  - Create monitoring dashboard with Grafana
- **Requirements**: Production readiness
- **Dependencies**: Core infrastructure
- **Estimated effort**: 2-3 weeks
- **Files to create**:
  - `internal/observability/`
  - `internal/metrics/`
  - `docker/monitoring/`

### 17.3 Performance optimization and caching

- **Status**: ❌ Not Started
- **Description**:
  - Implement Redis caching layer
  - Add database query optimization and indexing
  - Create API response caching with ETags
  - Implement connection pooling optimization
  - Add background job processing with queues
  - Create database read replicas support
  - Implement CDN integration for static assets
  - Add compression and minification
- **Requirements**: Scalability
- **Dependencies**: Core infrastructure
- **Estimated effort**: 2-3 weeks
- **Files to create**:
  - `internal/cache/`
  - `internal/jobs/`
  - `internal/infrastructure/redis/`

## 18. Advanced Tagging System Implementation

### 18.1 Implement comprehensive tagging system

- **Status**: ❌ Not Started
- **Description**:
  - Create tag parser for all @kthulu:\* tags
  - Implement dependency analyzer with automatic resolution
  - Add support for advanced tags (wrap, shadow, observable, microservice)
  - Create tag validation and conflict detection
  - Implement code generators based on tags
  - Add CLI commands for tag analysis and reporting
  - Create documentation and examples for all tag types
  - Add integration with existing build and generation scripts
- **Requirements**: Enhanced CLI capabilities
- **Dependencies**: All modules implemented
- **Estimated effort**: 3-4 weeks
- **Files to create**:
  - `internal/tags/parser.go`
  - `internal/tags/analyzer.go`
  - `internal/tags/generator.go`
  - `cmd/kthulu-analyzer/main.go`
  - `backend/docs/TAGGING_SYSTEM.md` ✅

### 18.2 Implement observability auto-generation

- **Status**: ❌ Not Started
- **Description**:
  - Parse @kthulu:observable tags and generate instrumentation
  - Create Prometheus metrics based on @kthulu:metrics tags
  - Generate OpenTelemetry tracing for tagged functions
  - Add structured logging with correlation IDs
  - Create Grafana dashboards from business metrics
  - Implement alerting rules based on tagged operations
  - Add performance profiling for critical paths
  - Create observability documentation and runbooks
- **Requirements**: Production observability
- **Dependencies**: 18.1 (Tagging system)
- **Estimated effort**: 2-3 weeks
- **Files to create**:
  - `internal/observability/generator.go`
  - `internal/observability/metrics.go`
  - `internal/observability/tracing.go`
  - `monitoring/dashboards/`
  - `monitoring/alerts/`

## 19. Single Binary Deployment

### 19.1 Implement single binary fullstack deployment

- **Status**: ✅ Completed
- **Description**:
  - ✅ Created StaticHandler for serving frontend assets
  - ✅ Implemented SPA routing with catch-all handler
  - ✅ Added security features (directory traversal protection, cache headers)
  - ✅ Created StaticModule with proper Fx integration
  - ✅ Updated ModuleSetBuilder to include static module
  - ✅ Created build-fullstack.sh script for automated building
  - ✅ Updated Vite configuration for backend/public output
  - ✅ Added Makefile targets for fullstack builds
  - ✅ Created optimized Dockerfile.fullstack for production
  - ✅ Added docker-compose.prod.yml for production deployment
  - ✅ Created comprehensive deployment documentation
- **Requirements**: Simplified deployment, developer experience
- **Dependencies**: Core infrastructure, frontend setup
- **Estimated effort**: 1 week ✅
- **Files created**:
  - `internal/adapters/http/static_handler.go` ✅
  - `internal/modules/static.go` ✅
  - `scripts/build-fullstack.sh` ✅
  - `Dockerfile.fullstack` ✅
  - `docker-compose.prod.yml` ✅
  - `docs/SINGLE_BINARY_DEPLOYMENT.md` ✅

## 20. Documentation and Final Polish

### 20.1 Create comprehensive documentation

- **Status**: ❌ Not Started
- **Description**:
  - Update README with complete setup instructions
  - Create architecture documentation with diagrams
  - Add API documentation with examples
  - Create deployment guides for different environments
  - Add troubleshooting and FAQ sections
  - Create contribution guidelines and code standards
  - Add security best practices documentation
  - Create user guides for all modules
- **Requirements**: 2.1, 10.1
- **Dependencies**: All modules completed
- **Estimated effort**: 1-2 weeks
- **Files to create**:
  - `docs/`
  - Updated `README.md`
  - `CONTRIBUTING.md`
  - `SECURITY.md`

### 20.2 Final testing and quality assurance

- **Status**: ❌ Not Started
- **Description**:
  - Run comprehensive test suite with coverage analysis
  - Perform security audit and vulnerability scanning
  - Execute performance testing and load testing
  - Validate all CLI deconstructibility tags
  - Test all module combinations and configurations
  - Verify OpenAPI specification completeness
  - Validate TypeScript type generation
  - Create final deployment and smoke tests
- **Requirements**: Quality assurance
- **Dependencies**: All previous tasks
- **Estimated effort**: 1-2 weeks
- **Files to create**:
  - `scripts/qa-check.sh`
  - `scripts/security-audit.sh`
  - `scripts/performance-test.sh`
