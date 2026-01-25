# AI-Native Flows

Kthulu is designed from the ground up to be an **AI-Native Software Foundry**. It doesn't just "have AI features"; it embeds AI into the core development loop through a set of powerful commands and integrations.

## 1. Kthulu Coder (The Agent)

`kthulu coder` is the flagship AI experience. It launches a dedicated TUI (Text User Interface) agent that acts as an autonomous engineer.

```bash
kthulu coder
```

### How it works
1.  **Environment Integration**: It detects your environment and automatically configures a local MCP (Model Context Protocol) server.
2.  **Tool Access**: The agent has direct access to Kthulu CLI commands (`add module`, `generate`, etc.) as **tools**. It doesn't guess how to create files; it uses the framework's own generators to ensure compliance.
3.  **Visual Interface**: Powered by [Crush](https://github.com/charmbracelet/crush), it provides a split-pane interface for chat and context management.

### Capabilities
- **Scaffolding**: "Create a user module with a repository and service."
- **Refactoring**: "Move the user validation logic to a separate helper."
- **Debugging**: "Analyze the logs and fix the panic in the auth handler."

---

## 2. Generative Commands (`kthulu ai`)

For quick, targeted tasks, the `ai` command suite provides direct access to generation models without leaving your terminal.

### Code Generation

Generate code snippets or entire files based on natural language prompts.

```bash
# General code generation
kthulu ai "Add a Stripe payment webhook handler" --apply

# Context-aware generation (reads project structure)
kthulu ai "Create a middleware to validate JWT tokens" --context
```

### Behavior Driven Development (BDD)

Kthulu automates the BDD workflow by generating Feature files and Step definitions.

```bash
# 1. Generate a Gherkin feature file
kthulu ai gen-feature "User registration with email verification" --apply

# 2. Generate the Go step definitions (Godog)
kthulu ai gen-steps features/registration.feature --apply
```

### Intelligent Review & optimization

Run AI-powered audits on your codebase.

```bash
# Security & Performance Review
kthulu ai review --fix-security --fix-performance

# Targeted Optimization
kthulu ai optimize --target=memory --benchmark
```

The optimizer will:
1.  Run a baseline benchmark (if `--benchmark` is set).
2.  Analyze the code for bottlenecks.
3.  Rewrite the code to improve the target metric.
4.  Run a verification benchmark to confirm improvements.

---

## 3. Context Awareness

Kthulu's AI flows are **Context Aware**. When you run a command, the system:
1.  **Analyzes the Project**: Scans `go.mod`, directory structure, and module definitions.
2.  **Extracts Patterns**: Identifies architectural patterns (e.g., "Modular Monolith", "Clean Architecture").
3.  **Feeds the LLM**: Passes this high-level context to the AI model so generated code fits your specific project style.

## Configuration

AI features can be configured via environment variables or flags:

- `OPENAI_API_KEY`: For OpenAI models.
- `ANTHROPIC_API_KEY`: For Claude models.
- `GEMINI_API_KEY`: For Google Gemini models.
- `LITELLM_BASE_URL`: For local models (Ollama, LocalAI) via LiteLLM.

```bash
# Use a specific provider
kthulu ai "..." --provider=anthropic --model=claude-3-sonnet
```
