# Kthulu Framework Comparison & Feature Gap Analysis

This document outlines a comparative analysis between **Kthulu Go** and established enterprise frameworks like **NestJS**, **Spring Boot**, **Buffalo**, and **Laravel**. It identifies key feature gaps and architectural differences to guide future roadmap development.

## 1. Microservices Native Support (gRPC & Messaging)

### The Gap
While Kthulu excels at HTTP/REST APIs, it takes a more explicit approach to non-HTTP transports compared to the "abstraction-heavy" competitors.

*   **Competitors (NestJS / Spring Boot):**
    *   **Transport Agnostic:** Switching a service from REST to **gRPC** or consuming events from **Kafka/RabbitMQ** is often a matter of changing a configuration flag or decorator. The framework abstracts the underlying protocol.
    *   **Unified Interface:** A controller method looks the same whether it's triggered by an HTTP GET request or a Kafka message.

*   **Kthulu Current State:**
    *   Primarily **HTTP/REST centric**.
    *   **gRPC Support:** Available via `grpc-gateway` to expose services as both gRPC and REST, but requires defining `.proto` files explicitly.
    *   **Background Jobs:** First-class support for Redis-based jobs via `asynq`, but lacking a unified event bus abstraction for Kafka/NATS.

### Why it matters
For large-scale distributed systems, HTTP overhead is often too high. Kthulu prioritizes the **Modular Monolith** pattern, where internal function calls replace network RPCs, deferring the need for complex microservice transports until absolutely necessary.

---

## 2. Automated Admin Interface (The "Admin Panel")

### The Gap
Kthulu provides tools to scaffold admin interfaces but does not offer a runtime-dynamic admin panel like Django.

*   **Competitors (Django / Laravel Nova / Buffalo):**
    *   **Zero-Config Admin:** Frameworks like Django automatically generate a full CRUD UI (`/admin`) based on database models at runtime.
    *   **Rapid Prototyping:** Drastically reduces time-to-market for internal tools.

*   **Kthulu Current State:**
    *   **Scaffolded Admin:** The `kthulu add admin` command generates a full **Templ + HTMX** admin dashboard source code into your project.
    *   **Difference:** Unlike Django's runtime inspection, Kthulu generates the *code* for the admin panel, giving you full control to customize it, but requiring a build step.

### Why it matters
An out-of-the-box admin panel is a massive productivity booster. Kthulu's code-generation approach balances convenience with the type-safety and performance of compiled Go code.

---

## 3. "Smart" Migrations & ORM DX

### The Gap
Kthulu uses industry-standard tools (GORM, Goose), but adheres to Go's explicit nature rather than "magic" automations.

*   **Competitors (Rails / Laravel / Buffalo):**
    *   **Intelligent Generators:** CLI commands like `rails g migration AddStatusToOrders status:string` automatically generate the correct SQL/DSL timestamped files.
    *   **Rich DSL:** Migration files use a high-level Domain Specific Language (e.g., `t.string :name`) that abstracts the specific SQL dialect differences completely.

*   **Kthulu Current State:**
    *   **Explicit SQL/Go:** Migrations are often raw SQL or basic Go structs.
    *   **Manual Work:** Developers often need to write the specific migration logic manually, although scaffolding helps.

### Why it matters
Smoother migration workflows reduce friction during rapid iteration cycles. Kthulu encourages understanding the underlying SQL, preventing "ORM magic" performance pitfalls later.

---

## 4. High-Level WebSocket Gateways

### The Gap
Real-time capabilities in Kthulu are functional but "bare-metal" compared to the abstractions offered by competitors.

*   **Competitors (NestJS Gateways / ActionCable):**
    *   **Abstractions:** Concepts like **Rooms**, **Namespaces**, and **Broadcasts** are first-class citizens.
    *   **Scalability Adapters:** Built-in adapters (e.g., Redis IoAdapter) allow WebSockets to scale across multiple server instances effortlessly.

*   **Kthulu Current State:**
    *   Support exists via libraries (e.g., `nhooyr.io/websocket`), but logic for managing connections, rooms, and clustering must be implemented by the developer.

### Why it matters
For apps requiring chat, live notifications, or collaborative editing, high-level abstractions save weeks of development time. This is a planned area of improvement for Kthulu.

---

## 5. Dependency Injection: Magic vs. Explicit

### The Difference
This is an architectural choice rather than a strict "missing feature," but it impacts Developer Experience.

*   **Competitors (Spring Boot / NestJS / Angular):**
    *   **Annotation-Based:** Uses decorators (`@Injectable`, `@Autowired`) and reflection to "magically" wire dependencies.
    *   **Pros:** Extremely terse code. No "wiring" files.
    *   **Cons:** "Magic" behavior can be hard to debug; runtime overhead (in some languages).

*   **Kthulu Current State (Uber/fx):**
    *   **Explicit & Compile-Time:** Uses `fx.Provide` and `fx.Invoke`. The dependency graph is constructed explicitly in Go code.
    *   **Pros:** Type-safe, high performance, easy to trace (Jump to Definition works).
    *   **Cons:** Requires maintaining `module.go` files and manually registering providers (Kthulu's CLI automates this wiring during generation).

### Why it matters
Kthulu prioritizes **performance and explicitness** (The Go Way) over "magic," which aligns with the ecosystem but requires a shift in mindset for developers coming from Java/TS.
