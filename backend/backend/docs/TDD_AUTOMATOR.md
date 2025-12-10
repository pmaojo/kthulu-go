# 🤖 Kthulu TDD Loop Automator: AI-Driven BDD

## 1. Vision
The **TDD Loop Automator** transforms Kthulu Go from a static scaffolding tool into a dynamic **Software Foundry**. It uses AI to automate the Test-Driven Development (TDD) cycle using Behavior-Driven Development (BDD) principles.

Instead of writing code first, the user describes **behavior**. Kthulu's AI Agents then iterate through the **Red-Green-Refactor** cycle automatically.

---

## 2. Architecture

The system leverages existing Kthulu components:

*   **Frontend (UX):** React-based "Behavior Lab" interface for interacting with features.
*   **Backend (Engine):** Go-based CLI utilizing `godog` for BDD execution.
*   **Control Plane (Integration):** MCP (Model Context Protocol) Server exposing CLI commands to AI.
*   **Intelligence (Brain):** `kthulu ai` command wrapping LLMs (Gemini/OpenAI) with project context awareness.

```mermaid
graph TD
    User[User / PO] -->|Natural Language| Frontend[React Frontend]
    Frontend -->|MCP Request| MCP[MCP Server]
    MCP -->|Executes| CLI[Kthulu CLI]
    CLI -->|Generates| Feature[.feature File]
    CLI -->|Generates| Steps[Step Definitions .go]
    CLI -->|Runs| Godog[Godog Runner]
    Godog -->|JSON Output| Frontend
    Frontend -->|Status (Red/Green)| User
```

---

## 3. The Automation Workflow

### Phase A: Spec Generation (AI as Business Analyst)
**Goal:** Define *what* the system should do.

1.  **Input:** User types: *"I want users to receive a welcome email upon registration."*
2.  **Process:**
    *   AI analyzes the `domain` layer (e.g., `User` struct).
    *   AI generates a Gherkin feature file.
3.  **Output (`backend/features/signup_notification.feature`):**
    ```gherkin
    Feature: User Signup Notification
      Scenario: Successful registration
        Given a new user with email "alice@example.com"
        When the user registers successfully
        Then a welcome email should be sent to "alice@example.com"
    ```

### Phase B: Step Implementation (AI as SDET)
**Goal:** Make the test executable (but failing).

1.  **Status:** Frontend shows 🟡 **Undefined Steps**.
2.  **Action:** AI reads the `.feature` file and the `godog_test.go` context.
3.  **Process:**
    *   AI identifies missing step definitions.
    *   AI generates Go code mapping regex to test logic.
4.  **Output (`backend/features/signup_steps_test.go`):**
    ```go
    ctx.Step(`^a welcome email should be sent to "([^"]*)"$`, func(email string) error {
        // AI injects mock assertion logic
        if !mockMailer.WasCalled("SendWelcome", email) {
            return fmt.Errorf("email not sent")
        }
        return nil
    })
    ```
5.  **Status:** Frontend shows 🔴 **Failing Test** (Red).

### Phase C: Code Realization (AI as Developer)
**Goal:** Make the test pass.

1.  **Input:** User clicks **"Fix Code"**.
2.  **Process:**
    *   AI analyzes the failure: *"email not sent"*.
    *   AI locates the `Service` layer (`internal/adapters/http/modules/users/service`).
    *   AI injects the missing business logic.
3.  **Output (Code Change):**
    ```go
    // In UserService.Register
    func (s *Service) Register(user *User) error {
        // ... save user ...
        s.notifier.SendWelcome(user.Email) // <--- AI adds this
        return nil
    }
    ```
4.  **Status:** Frontend shows 🟢 **Passing Test** (Green).

---

## 4. Integration Guide

### 4.1 Frontend (`frontend/src/`)
*   **New Component:** `BehaviorLab.tsx`
    *   Split view: Chat (Left) | Feature/Code Viewer (Right).
    *   Status indicators for Scenarios.
*   **API Service:** Enhance `kthuluApi.ts` to parse `go test -json` output.

### 4.2 CLI (`backend/backend/cmd/kthulu-cli/`)
*   **Enhance `ai` command:**
    *   Support target-specific generation: `kthulu ai gen-feature` vs `kthulu ai gen-steps`.
    *   Implement the `--apply` flag to write AI-generated code to disk (currently TODO).
*   **Enhance `test` command:**
    *   Ensure `godog` output can be captured as JSON for the frontend.

### 4.3 MCP (`backend/backend/internal/adapters/mcp/`)
*   **New Tools:**
    *   `list_features`: Return list of `.feature` files.
    *   `read_feature`: Return content of a feature.
    *   `run_scenario`: Execute specific scenario by line number or tag.

### 4.4 Terminal UI (TUI) Integration
For developers who prefer the terminal, the `kthulu debug` command will be enhanced with a **Test Watch** mode.

*   **Command:** `kthulu debug --test-watch`
*   **New Tab:** **"🧪 Tests"** tab alongside "HTTP" and "Database".
*   **Real-time Feedback:**
    *   Watches for file changes in `.feature` or `_test.go` files.
    *   Automatically triggers `godog` execution in the background.
    *   Parses JSON output to render a Red/Green progress bar and list failing scenarios directly in the terminal.

---

## 5. Future Roadmap
*   **Self-Healing Tests:** If a UI change breaks a test, AI automatically updates the step definition.
*   **Edge Case Discovery:** AI proposes additional Scenarios (Negative testing) automatically.
*   **Performance BDD:** Gherkin steps for latency requirements (*"Then response time should be < 50ms"*).
