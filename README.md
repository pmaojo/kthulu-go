# Kthulu Go — The AI-Native Software Foundry

**Kthulu Go** is not just a framework—it's an **Intelligent Software Foundry**. It combines a robust Modular Monolith architecture with a powerful **Generation Engine** driven by AI and the Model Context Protocol (MCP).

Designed for platform engineering teams and AI-native developers, Kthulu automates the heavy lifting of software design, allowing you to focus on business logic.

![Kthulu Architecture](https://kthulu.dev/assets/architecture-diagram.png)

## 🚀 Core Pillars

### 1. The Generation Engine (DSL)
Infrastructure as Code is standard. Kthulu brings you **Architecture as Code**.
Define your system's blueprint using `kthulu plan`, and let the engine scaffold a production-ready Modular Monolith.

[👉 Learn more about Project Blueprints](./docs/DSL.md)

### 2. AI-Native Workflow
Kthulu embeds AI into the developer loop. It doesn't just autocomplete code; it understands your project context, audits your security, and writes your tests.

- **Kthulu Coder**: An autonomous TUI agent that lives in your terminal.
- **Generative Commands**: `kthulu ai gen-feature`, `kthulu ai review`, `kthulu ai optimize`.

[👉 Deep dive into AI Flows](./docs/AI_FLOWS.md)

### 3. Modular Monolith Architecture
We enforce a "Vertical Slice" architecture (GTH Stack: Go + Templ + HTMX). Code is organized by **Feature**, not technical layer, ensuring high cohesion and scalability.

### 4. MCP Native
Kthulu exposes its entire CLI as a **Model Context Protocol (MCP)** server. This means any MCP-compliant agent (Claude Desktop, Cursor, etc.) can "drive" Kthulu to build software for you.

[👉 Integrating with MCP](./docs/MCP.md)

---

## ⚡ Quick Start

### Installation

```bash
# Build from source
go build -o kthulu ./cmd/kthulu/main.go

# Add to PATH
export PATH="$(pwd):$PATH"
```

### 1. Plan Your Architecture
Don't start with an empty folder. Start with a plan.

```bash
# Generate a blueprint for an e-commerce app
kthulu plan my-shop --template=ecommerce --features=payment,cart
```

### 2. Scaffold the Project
Turn the blueprint into code.

```bash
# Create the project straight from the blueprint
kthulu create my-shop --from-plan=kthulu-plan.yaml

# Or use flags directly
kthulu create my-shop --features=payment,cart
```

### 3. Launch the AI Agent
Need to add a feature? Let the agent handle the boilerplate.

```bash
cd my-shop
kthulu coder
```

---

## 🛠 Feature Highlights

### 🧠 Intelligent Coding
Forget copy-pasting from ChatGPT. Kthulu's AI is context-aware.

```bash
# Add a Stripe webhook handler, context-aware
kthulu ai "Add a Stripe payment webhook handler" --apply

# Generate BDD tests
kthulu ai gen-feature "User checkout flow" --apply
```

### ✅ Declarative Validation
Declare rules in the blueprint; get a generated `Validate()` method, service-level enforcement, and 422 responses with field-level error maps.

```yaml
modules:
  product:
    fields:
      - name:string:required,min=2
      - price:int:required,min=1
      - status:string:oneof=draft|active|archived
```

### ⚙️ Background Jobs & Scheduler
Add the `queues` feature and get a database-backed job runtime — worker pool, exponential-backoff retries, dead letters, and recurring schedules — with zero extra infrastructure. The `mail` and `storage` features generate env-configured drivers (SMTP/log mailer, local-disk storage) injected via Fx.

```go
q.Enqueue(jobs.TypeWelcomeEmail, jobs.WelcomeEmail{UserID: 42})
q.Every(time.Hour, jobs.TypeHeartbeat, nil) // cron equivalent
```

### 🔮 Interactive Console
The `rails console` / `artisan tinker` equivalent — inspect and manipulate your project's database with auto-discovered connection settings.

```bash
kthulu console
kthulu> list Products
kthulu> create Products name=Widget price=10
kthulu> sql SELECT name, price FROM Products
```

### 🧬 Smart Migrations
`kthulu migrate diff` compares your entity structs against the live database and generates the SQL migration for the gap — additive only, destructive changes are reported but never applied.

```bash
kthulu migrate diff --dry-run
kthulu migrate diff --name add_reviews
```

### 🥒 Behavior Driven Development (BDD)
First-class support for Cucumber/Gherkin. Define *behavior* first.

```bash
# Run all feature tests
kthulu bdd run
```

### 🛡️ Enterprise Compliance & Security
Built-in auditing tools for security and compliance (SOX, GDPR, PCI).

```bash
# Run a security and compliance audit
kthulu audit --security --compliance=gdpr
```

### ☁️ Zero-Config Deployment
Deploy to any cloud provider (AWS, GCP, Azure, K8s) without writing Terraform.

```bash
kthulu deploy --cloud=aws --scale=auto
```

---

## 📚 Documentation

- **[AI Flows & Capabilities](./docs/AI_FLOWS.md)**: Master the `coder` and `ai` commands.
- **[Project Blueprints (DSL)](./docs/DSL.md)**: Learn how to define Architecture as Code.
- **[Model Context Protocol (MCP)](./docs/MCP.md)**: Connect Kthulu to Claude and other agents.
- **[Framework Comparison](./docs/FRAMEWORK_COMPARISON.md)**: Why Kthulu vs. NestJS or Spring Boot?

## 📦 Project Structure

```
my-app/
├── kthulu-plan.yaml     # Architecture Blueprint
├── cmd/server/          # Entrypoint
├── features/            # BDD Feature Files (.feature)
├── internal/
│   ├── modules/         # Vertical Slices (User, Billing, etc.)
│   ├── views/           # GTH Frontend (Templ + HTMX)
│   └── infrastructure/  # Shared Kernels
└── ...
```

## 🤝 Contributing

We love contributions! Please read our [CONTRIBUTING.md](./CONTRIBUTING.md) to get started.

## 📄 License

MIT — see [LICENSE](./LICENSE)
