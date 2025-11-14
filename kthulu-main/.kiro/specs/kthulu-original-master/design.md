# Design Document

## Overview

The `kthulu-original` master project implements a complete enterprise-grade monolith with hexagonal architecture, serving as the reference template for the `kthulu` CLI scaffolder. The design emphasizes modularity, type safety, and "deconstructibility" - allowing selective module extraction for generated projects.

## Architecture

### Hexagonal Architecture Layers

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐    ┌──────────────────┐
│    Adapters     │───▶│    Use Cases     │───▶│   Repository    │───▶│  Infrastructure  │
│   (HTTP/CLI)    │    │   (Business)     │    │  (Interfaces)   │    │   (Database)     │
└─────────────────┘    └──────────────────┘    └─────────────────┘    └──────────────────┘
```

**Dependency Flow:** Adapters → Use Cases → Repository Interfaces → Infrastructure Implementations

### Project Structure

```
/kthulu-original/
├── backend/
│   ├── cmd/service/              # Application entry point with Fx wiring
│   ├── internal/
│   │   ├── core/                 # Cross-cutting concerns (config, logger, db)
│   │   ├── modules/              # Business modules (auth, user, org, etc.)
│   │   ├── adapters/http/        # HTTP handlers and middleware
│   │   ├── usecase/              # Business logic coordinators
│   │   ├── repository/           # Data access interfaces
│   │   └── infrastructure/db/    # Database implementations
│   └── migrations/               # Database schema evolution
├── frontend/src/                 # Vite + React + TypeScript application
├── scripts/                      # Build and generation scripts
├── api/                          # Generated OpenAPI specifications
└── docker-compose.yml            # Development environment
```

## Components and Interfaces

### Core Infrastructure

#### Configuration Management
- **Location:** `internal/core/config.go`
- **Purpose:** Centralized configuration loading from environment variables
- **Dependencies:** `godotenv` for .env file support
- **Tag:** `// @kthulu:core`

```go
type Config struct {
    Database DatabaseConfig
    Server   ServerConfig
    JWT      JWTConfig
    SMTP     SMTPConfig
}
```

#### Logging System
- **Location:** `internal/core/logger.go`
- **Purpose:** Structured logging with Zap
- **Configuration:** Different levels for development/production
- **Tag:** `// @kthulu:core`

#### Database Connection
- **Location:** `internal/core/database.go`
- **Purpose:** GORM connection management with automatic migrations
- **Features:** Connection pooling, health checks, migration runner
- **Tag:** `// @kthulu:core`

### Module Architecture

Each module follows a consistent structure:

```
internal/modules/<module_name>/
├── module.go           # Fx module definition and registration
├── entity.go           # Domain entities and value objects
├── repository.go       # Repository interface definition
├── usecase.go          # Business logic implementation
└── handler.go          # HTTP handlers and DTOs
```

#### Module Registration Pattern

```go
// @kthulu:module:<module_name>
func Module() fx.Option {
    return fx.Module("<module_name>",
        fx.Provide(
            NewRepository,
            NewUseCase,
            NewHandler,
        ),
        fx.Invoke(RegisterRoutes),
    )
}
```

### Core Modules

#### Authentication Module (`auth`)
- **Endpoints:** POST /auth/register, /auth/login, /auth/confirm, /auth/refresh, /auth/logout
- **Features:** JWT token management, email confirmation, refresh token rotation
- **Dependencies:** notifier (for email), user (for profile data)
- **Database Tables:** users, refresh_tokens

#### User Module (`user`)
- **Endpoints:** GET /users/me, PATCH /users/me
- **Features:** Profile management, user preferences
- **Dependencies:** auth (for authentication)
- **Database Tables:** users (shared with auth)

#### Access Control Module (`access`)
- **Features:** Role-based access control (RBAC), middleware for route protection
- **Middleware:** X-Role-Scope header validation
- **Dependencies:** auth (for user context)
- **Database Tables:** roles, user_roles, permissions

#### Notification Module (`notifier`)
- **Purpose:** Event-driven email notifications
- **Implementation:** Console output for development, SMTP for production
- **Interface:** Pluggable notification providers
- **Dependencies:** None (pure service)

### ERP-lite Modules

#### Organization Module (`org`)
- **Features:** Multi-tenant organizations, invitations, domain management
- **Dependencies:** auth, user
- **Database Tables:** organizations, organization_users, invitations

#### Contacts Module (`contacts`)
- **Features:** Customer/supplier management, contact information
- **Dependencies:** org (for tenant isolation)
- **Database Tables:** contacts, contact_addresses, contact_phones

#### Products Module (`products`)
- **Features:** Product catalog, pricing, variants, tax classifications
- **Dependencies:** org (for tenant isolation)
- **Database Tables:** products, product_variants, product_prices

#### Invoices Module (`invoices`)
- **Features:** Invoice generation, tax calculation, payment tracking, PDF export
- **Dependencies:** contacts, products
- **Database Tables:** invoices, invoice_items, payments

#### Veri*Factu Module (`verifactu`) - Spanish Tax Compliance
- **Features:** Spanish tax compliance for invoice verification (RD 1007/2023 RRSIF)
- **Modes:** Real-time AEAT submission or queued with digital signature
- **Dependencies:** invoices (extends invoice functionality)
- **Database Tables:** verifactu_records, verifactu_events, verifactu_signatures
- **Integration:** QR codes, legal legends, XML/JSON structured records
- **Compliance:** AEAT real-time submission, integrity verification, audit trails

#### Inventory Module (`inventory`)
- **Features:** Stock management, warehouse operations, movement tracking
- **Dependencies:** products
- **Database Tables:** warehouses, inventory_items, stock_movements

#### Calendar Module (`calendar`)
- **Features:** Appointment scheduling, event management, availability tracking
- **Dependencies:** user, org
- **Database Tables:** events, appointments, availability_slots

## Data Models

### Core Entities

```go
// @kthulu:core
type User struct {
    ID          uint      `gorm:"primaryKey"`
    Email       string    `gorm:"uniqueIndex" validate:"required,email"`
    PasswordHash string   `validate:"required"`
    ConfirmedAt *time.Time
    RoleID      uint
    Role        Role      `gorm:"foreignKey:RoleID"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// @kthulu:core
type Role struct {
    ID          uint   `gorm:"primaryKey"`
    Name        string `gorm:"uniqueIndex" validate:"required"`
    Description string
    Permissions []Permission `gorm:"many2many:role_permissions"`
}

// @kthulu:core
type RefreshToken struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `validate:"required"`
    Token     string    `gorm:"uniqueIndex" validate:"required"`
    ExpiresAt time.Time `validate:"required"`
    CreatedAt time.Time
}
```

### Module-Specific Entities

Each ERP-lite module defines its own entities following domain-driven design principles, with proper foreign key relationships and validation rules.

## Error Handling

### Error Types
- **Domain Errors:** Business rule violations (e.g., `ErrUserAlreadyExists`)
- **Infrastructure Errors:** Database, network, external service failures
- **Validation Errors:** Input validation failures using `validator/v10`

### Error Response Format
```go
type ErrorResponse struct {
    Error   string            `json:"error"`
    Code    string            `json:"code"`
    Details map[string]string `json:"details,omitempty"`
}
```

## Testing Strategy

### Contract Testing
- **Location:** `internal/contracts/`
- **Purpose:** Verify repository implementations satisfy their interfaces
- **Pattern:** Interface assignment tests with httptest snapshots
- **Coverage:** All repository interfaces must have contract tests

```go
// @kthulu:core
func TestUserRepositoryContract(t *testing.T) {
    var _ repository.UserRepository = (*db.UserRepository)(nil)
    // Additional behavioral tests...
}
```

### Integration Testing
- **Database:** Use testcontainers for PostgreSQL integration tests
- **HTTP:** Use httptest for API endpoint testing
- **Modules:** Test module interactions through use case tests

### Frontend Testing
- **Unit Tests:** Vitest for utility functions and hooks
- **Component Tests:** React Testing Library for UI components
- **E2E Tests:** Playwright for critical user journeys
- **Coverage Target:** Minimum 70% for frontend, 80% for backend

## Type Safety and Code Generation

### OpenAPI Generation
- **Script:** `scripts/openapi.sh`
- **Tool:** `oapi-codegen` with custom templates
- **Output:** `api/openapi.yaml` with complete API specification
- **Automation:** Runs in CI/CD pipeline and pre-commit hooks

### TypeScript Generation
- **Script:** `scripts/gen-types.sh`
- **Process:** OpenAPI → TypeScript interfaces + Zod schemas
- **Output:** `frontend/src/types/kthulu-api.ts`
- **Integration:** Automatic import in frontend services and components

### Build Process
```bash
# Development workflow
make openapi      # Generate OpenAPI spec from Go structs
make gen-types    # Generate TypeScript types from OpenAPI
make dev          # Start development environment
make test         # Run all tests
```

## Frontend Architecture

### Technology Stack
- **Build Tool:** Vite 5+ for fast development and building
- **Framework:** React 18 with TypeScript 5.4+
- **Styling:** Tailwind CSS 3.4+ for utility-first styling
- **State Management:** Zustand for global state (one slice per bounded context)
- **Data Fetching:** TanStack Query v5 for server state management
- **HTTP Client:** Axios with automatic token refresh interceptors

### Component Organization
```
src/
├── components/
│   ├── layouts/     # Page layouts (AdminLayout, PublicLayout)
│   ├── shared/      # Reusable UI components
│   └── views/       # Page components organized by domain
├── hooks/           # Custom React hooks
├── services/        # API service functions
├── stores/          # Zustand store definitions
└── types/           # Generated TypeScript types
```

### Authentication Flow
1. User submits credentials via `/login` form
2. Frontend calls `POST /auth/login` API endpoint
3. Backend validates credentials and returns JWT access + refresh tokens
4. Axios interceptor automatically adds Authorization header to subsequent requests
5. On 401 response, interceptor attempts token refresh via `/auth/refresh`
6. Failed refresh redirects user to login page

## Deployment and Development

### Docker Configuration
- **Backend:** Multi-stage build with Go 1.22, optimized binary
- **Frontend:** Vite build served by nginx in production
- **Database:** PostgreSQL 15 with persistent volume
- **Development:** Hot reload for both backend (air) and frontend (Vite HMR)

### Environment Configuration
- **Development:** `.env` file with local settings
- **Production:** Environment variables with secure defaults
- **Database:** Automatic migration on startup
- **Logging:** Structured JSON logs in production, human-readable in development

## CLI Deconstructibility Features

### Advanced File Tagging System

The Kthulu framework uses a comprehensive tagging system that goes beyond simple module identification:
- **Core Files:** `// @kthulu:core` for framework essentials
- **Module Files:** `// @kthulu:module:<name>` for module-specific code
- **Generated Files:** `// @kthulu:generated` for auto-generated content

### Complete ERP-lite Module Architecture

#### Inventory Module (`inventory`)
- **Features:** Multi-warehouse management, stock movements, inventory valuation, reorder points
- **Dependencies:** products (for stock tracking)
- **Database Tables:** warehouses, inventory_items, stock_movements, stock_adjustments
- **Integration:** Real-time stock updates, low-stock alerts, FIFO/LIFO/Average costing

#### Calendar Module (`calendar`)
- **Features:** Appointment scheduling, availability management, recurring events, notifications
- **Dependencies:** user, org (for calendar ownership and sharing)
- **Database Tables:** calendars, events, appointments, availability_slots, bookings
- **Integration:** Timezone support, conflict detection, reminder notifications

### Advanced Architecture Features

#### Contract Testing Framework
- **Location:** `internal/contracts/`
- **Purpose:** Verify interface compliance and behavioral contracts
- **Coverage:** Repository interfaces, HTTP endpoints, use case behaviors
- **Tools:** httptest, testcontainers, golden files for snapshot testing

#### Frontend Architecture Enhancement
- **State Management:** Zustand with module-specific slices
- **Data Fetching:** TanStack Query with optimistic updates and caching
- **UI Components:** Design system with Tailwind CSS and accessibility support
- **Testing:** React Testing Library with component and integration tests

#### Observability Stack
- **Metrics:** Prometheus with custom business metrics
- **Tracing:** OpenTelemetry with distributed request tracking
- **Logging:** Structured JSON logs with correlation IDs
- **Monitoring:** Health checks, performance profiling, error tracking

#### Security Enhancements
- **Authentication:** OAuth2, 2FA, session management, device tracking
- **Authorization:** Enhanced RBAC with resource-level permissions
- **Audit:** Security event logging and compliance reporting
- **Protection:** Rate limiting, account lockout, password policies

### Module Independence
- Each module can be enabled/disabled via fx.Option inclusion
- Frontend components are organized by module for selective copying
- Database migrations are tagged by module for selective application
- OpenAPI specifications include module tags for subset generation
- Contract tests verify module boundaries and interfaces

### Advanced Tagging Capabilities

#### **Planned Advanced Tags**
- **Extensibility:** `// @kthulu:wrap` (safe extension points), `// @kthulu:shadow` (complete override)
- **Observability:** `// @kthulu:observable` (auto-instrumentation), `// @kthulu:metrics:<type>`
- **Architecture:** `// @kthulu:microservice` (microservice candidates), `// @kthulu:dependency:<modules>`
- **Security:** `// @kthulu:security:<level>` (access control), `// @kthulu:audit` (audit logging)
- **Quality:** `// @kthulu:deprecated` (obsolete code), `// @kthulu:experimental` (beta features)
- **Generation:** `// @kthulu:cli:generator` (code templates), `// @kthulu:cli:config` (user prompts)

#### **Tag Parser Architecture**
```go
// @kthulu:core
type Tag struct {
    Type       string            // "core", "module", "observable", etc.
    Value      string            // Primary value (module name, etc.)
    Attributes map[string]string // Additional attributes
    File       string            // Source file
    Line       int               // Line number
}
```

### Selective Generation
The CLI will use these design patterns to:
1. Copy only selected modules based on user choices
2. Generate appropriate fx.Module configurations
3. Create subset OpenAPI specifications
4. Build frontend with only required components
5. Generate docker-compose with only needed services
6. Include relevant tests and documentation
7. Configure monitoring and security features
8. Generate observability instrumentation based on tags
9. Create extension points and customization hooks
10. Validate security and compliance requirements
## Sp
anish Tax Compliance (Veri*Factu Module)

### Overview
The Veri*Factu module provides optional compliance with Spanish tax regulations (RD 1007/2023 RRSIF) for invoice verification. It integrates seamlessly with the existing invoice module without breaking the clean architecture.

### Architecture Integration
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Invoice Module │───▶│ Veri*Factu Module│───▶│  AEAT Service   │
│   (Core ERP)    │    │   (Compliance)   │    │  (External)     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Compliance Features
- **Structured Records:** XML/JSON generation according to RRSIF specifications
- **Dual Modes:** Real-time AEAT submission or queued with digital signatures
- **QR Integration:** Automatic QR code generation with verification links
- **Audit Trail:** Complete event logging for compliance verification
- **Multi-terminal Support:** Independent operation across multiple terminals
- **Retry Mechanisms:** Automatic retry with incident logging for network failures

### Data Models
```go
// @kthulu:module:verifactu
type VeriFactuRecord struct {
    ID               uint                 `json:"id"`
    InvoiceID        uint                 `json:"invoiceId"`
    OrganizationID   uint                 `json:"organizationId"`
    RecordType       VeriFactuRecordType  `json:"recordType"` // alta, baja, incident
    StructuredData   string               `json:"structuredData"` // XML/JSON
    Hash             string               `json:"hash"`
    Signature        string               `json:"signature,omitempty"`
    QRCode           string               `json:"qrCode"`
    SubmissionStatus VeriFactuStatus      `json:"submissionStatus"`
    SubmittedAt      *time.Time           `json:"submittedAt,omitempty"`
    CreatedAt        time.Time            `json:"createdAt"`
}

type VeriFactuEvent struct {
    ID           uint                `json:"id"`
    RecordID     uint                `json:"recordId"`
    EventType    VeriFactuEventType  `json:"eventType"`
    Description  string              `json:"description"`
    UserID       uint                `json:"userId"`
    CreatedAt    time.Time           `json:"createdAt"`
}
```

### Configuration
```bash
# Enable Veri*Factu module
MODULES=invoice,verifactu

# Veri*Factu specific configuration
VERIFACTU_MODE=real-time  # or 'queued'
VERIFACTU_AEAT_ENDPOINT=https://sede.agenciatributaria.gob.es/...
VERIFACTU_CERTIFICATE_PATH=/path/to/certificate.p12
VERIFACTU_CERTIFICATE_PASSWORD=secret
VERIFACTU_ORGANIZATION_NIF=12345678A
```

### Integration Points
- **Invoice Creation:** Automatic record generation on invoice save
- **PDF Generation:** QR code and legal legend injection
- **Event System:** Audit trail for all compliance operations
- **Error Handling:** Graceful degradation when AEAT is unavailable

### Module Independence
- **Optional Activation:** Can be enabled/disabled via configuration
- **Zero Coupling:** Invoice module works independently without Veri*Factu
- **Event-Driven:** Uses domain events for loose coupling
- **Extensible:** Can be extended for other tax jurisdictions (Portugal, Italy, etc.)