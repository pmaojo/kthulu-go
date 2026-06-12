package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// ---------------------------------------------------------------------------
// generate_method tool
// ---------------------------------------------------------------------------

// GenerateMethodArgs defines arguments for the generate_method tool.
type GenerateMethodArgs struct {
	Module      string `json:"module"       jsonschema:"required,description=Module name (e.g. 'order')."`
	MethodName  string `json:"method_name"  jsonschema:"required,description=Method name in CamelCase (e.g. 'CalculateTax')."`
	Description string `json:"description"  jsonschema:"required,description=What the method should do in plain language. E.g. 'Calculate 21% VAT on total_cents, exempt if user.country is in EU list'."`
	DryRun      bool   `json:"dry_run,omitempty" jsonschema:"description=If true, return the generated code without writing it to the file."`
}

func generateMethodTool(executor CommandExecutor, workingDir string) RegisteredTool {
	return RegisteredTool{
		Name: "generate_method",
		Description: "Generate a new service method for a module using AI (requires OPENAI_API_KEY) or a typed stub. " +
			"Writes directly to the module's service.go file after showing a preview.",
		Handler: func(ctx context.Context, args GenerateMethodArgs) (*mcp_golang.ToolResponse, error) {
			if strings.TrimSpace(args.Module) == "" {
				return nil, fmt.Errorf("module is required")
			}
			if strings.TrimSpace(args.MethodName) == "" {
				return nil, fmt.Errorf("method_name is required")
			}
			if strings.TrimSpace(args.Description) == "" {
				return nil, fmt.Errorf("description is required")
			}

			dir := resolveWorkdir(workingDir)
			servicePath := filepath.Join(dir, "internal", args.Module, "service.go")

			data, err := os.ReadFile(servicePath)
			if err != nil {
				return nil, fmt.Errorf("cannot read %s: %w", servicePath, err)
			}
			fileContent := string(data)

			// Parse the file to extract service struct name
			structName, err := parseServiceStructName(servicePath)
			if err != nil {
				// Fall back to a title-cased module name
				structName = titleCase(args.Module) + "Service"
			}

			// Check if method already exists
			if methodExists(servicePath, args.MethodName) {
				return nil, fmt.Errorf("method %s already exists in %s", args.MethodName, servicePath)
			}

			var generatedCode string

			apiKey := os.Getenv("OPENAI_API_KEY")
			if apiKey != "" {
				generatedCode, err = generateMethodViaOpenAI(ctx, apiKey, fileContent, structName, args.MethodName, args.Description)
				if err != nil {
					// Fall back to stub on AI error
					generatedCode = buildMethodStub(structName, args.MethodName, args.Description)
				}
			} else {
				generatedCode = buildMethodStub(structName, args.MethodName, args.Description)
			}

			if args.DryRun {
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(
					fmt.Sprintf("// DRY RUN — would append to %s\n\n%s", servicePath, generatedCode))), nil
			}

			// Append to service.go
			if err := appendToFile(servicePath, "\n"+generatedCode+"\n"); err != nil {
				return nil, fmt.Errorf("failed to append to %s: %w", servicePath, err)
			}

			// Verify it compiles
			if buildErr := goBuild(ctx, dir, args.Module); buildErr != nil {
				// Roll back
				_ = os.WriteFile(servicePath, data, 0o644)
				return nil, fmt.Errorf("generated code does not compile (rolled back):\n%w\n\nGenerated code was:\n%s", buildErr, generatedCode)
			}

			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(
				fmt.Sprintf("Method %s appended to %s\n\n%s", args.MethodName, servicePath, generatedCode))), nil
		},
	}
}

// parseServiceStructName finds the first exported struct ending in "Service" in the file.
func parseServiceStructName(path string) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return "", err
	}
	symbols, _, err := parseFileSymbols(filepath.Dir(path), path)
	if err != nil {
		return "", err
	}
	_ = f
	for _, s := range symbols {
		if s.Kind == "struct" && strings.HasSuffix(s.Name, "Service") {
			return s.Name, nil
		}
	}
	return "", fmt.Errorf("no *Service struct found in %s", path)
}

// methodExists returns true if a method with the given name already exists in the file.
func methodExists(path, methodName string) bool {
	symbols, _, err := parseFileSymbols(filepath.Dir(path), path)
	if err != nil {
		return false
	}
	for _, s := range symbols {
		if s.Kind == "method" && s.Name == methodName {
			return true
		}
	}
	return false
}

// buildMethodStub generates a typed stub method.
func buildMethodStub(structName, methodName, description string) string {
	receiver := strings.ToLower(string([]rune(structName)[0]))
	return fmt.Sprintf(
		`// %s implements: %s
func (%s *%s) %s(ctx context.Context) error {
	// TODO: implement %s
	// Description: %s
	return fmt.Errorf("not implemented")
}
`,
		methodName,
		description,
		receiver,
		structName,
		methodName,
		methodName,
		description,
	)
}

// openAIRequest is a minimal OpenAI chat completions request.
type openAIRequest struct {
	Model       string           `json:"model"`
	Messages    []openAIMessage  `json:"messages"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float32          `json:"temperature,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func generateMethodViaOpenAI(ctx context.Context, apiKey, fileContent, structName, methodName, description string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert Go developer. Here is an existing service file:

%s

Generate ONLY the new method body as valid Go code for the struct %s.
Method name: %s
What it should do: %s

Requirements:
- Match the patterns and conventions in the existing file exactly.
- Include the complete method signature and body.
- Include any helper types needed.
- Do NOT repeat or duplicate any existing code from the file.
- Do NOT include package declaration or imports (they are already present).
- Output ONLY raw Go code with no markdown fences.
`,
		fileContent, structName, methodName, description)

	req := openAIRequest{
		Model: "gpt-4o-mini",
		Messages: []openAIMessage{
			{Role: "system", Content: "You are an expert software architect and Go code generator. Output only valid Go code."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1024,
		Temperature: 0.2,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("openai error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in openai response")
	}

	code := result.Choices[0].Message.Content
	// Strip markdown fences if the model adds them despite instructions
	code = stripMarkdownFences(code)
	return code, nil
}

func stripMarkdownFences(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			continue
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n")) + "\n"
}

// appendToFile appends content to the given file.
func appendToFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// goBuild runs go build on the module package and returns any error.
func goBuild(ctx context.Context, workingDir, module string) error {
	pkg := fmt.Sprintf("./internal/%s/...", module)
	cmd := exec.CommandContext(ctx, "go", "build", pkg)
	cmd.Dir = workingDir
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, string(out))
	}
	return nil
}

// titleCase uppercases the first rune of s.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// ---------------------------------------------------------------------------
// add_hook tool
// ---------------------------------------------------------------------------

// AddHookArgs defines arguments for the add_hook tool.
type AddHookArgs struct {
	Module      string `json:"module"      jsonschema:"required,description=Module name."`
	Lifecycle   string `json:"lifecycle"   jsonschema:"required,description=Hook lifecycle: before_create, after_create, before_update, after_update, before_delete, after_delete."`
	Description string `json:"description" jsonschema:"required,description=What the hook should do. E.g. 'send welcome email via mail service' or 'validate that stock > 0'."`
}

// lifecycleToMethod maps lifecycle identifiers to GORM hook method names.
var lifecycleToMethod = map[string]string{
	"before_create": "BeforeCreate",
	"after_create":  "AfterCreate",
	"before_update": "BeforeUpdate",
	"after_update":  "AfterUpdate",
	"before_delete": "BeforeDelete",
	"after_delete":  "AfterDelete",
}

func addHookTool(executor CommandExecutor, workingDir string) RegisteredTool {
	return RegisteredTool{
		Name: "add_hook",
		Description: "Add a lifecycle hook to a module (before_create, after_create, before_update, after_update, before_delete, after_delete). " +
			"Hooks are implemented as GORM hooks on the model struct.",
		Handler: func(ctx context.Context, args AddHookArgs) (*mcp_golang.ToolResponse, error) {
			if strings.TrimSpace(args.Module) == "" {
				return nil, fmt.Errorf("module is required")
			}
			if strings.TrimSpace(args.Lifecycle) == "" {
				return nil, fmt.Errorf("lifecycle is required")
			}
			if strings.TrimSpace(args.Description) == "" {
				return nil, fmt.Errorf("description is required")
			}

			methodName, ok := lifecycleToMethod[args.Lifecycle]
			if !ok {
				validKeys := make([]string, 0, len(lifecycleToMethod))
				for k := range lifecycleToMethod {
					validKeys = append(validKeys, k)
				}
				return nil, fmt.Errorf("unknown lifecycle %q; valid values: %s", args.Lifecycle, strings.Join(validKeys, ", "))
			}

			dir := resolveWorkdir(workingDir)
			modelPath := filepath.Join(dir, "internal", args.Module, "model.go")

			data, err := os.ReadFile(modelPath)
			if err != nil {
				return nil, fmt.Errorf("cannot read %s: %w", modelPath, err)
			}

			// Parse model struct name
			modelStructName, err := parseModelStructName(modelPath)
			if err != nil {
				modelStructName = titleCase(args.Module)
			}

			// Check if hook already exists
			if methodExists(modelPath, methodName) {
				return nil, fmt.Errorf("hook %s already exists on %s in %s", methodName, modelStructName, modelPath)
			}

			hookCode := buildHookCode(modelStructName, methodName, args.Description)

			// Append to model.go
			if err := appendToFile(modelPath, "\n"+hookCode+"\n"); err != nil {
				return nil, fmt.Errorf("failed to append to %s: %w", modelPath, err)
			}

			// Verify it compiles
			if buildErr := goBuild(ctx, dir, args.Module); buildErr != nil {
				// Roll back
				_ = os.WriteFile(modelPath, data, 0o644)
				return nil, fmt.Errorf("hook code does not compile (rolled back):\n%w\n\nHook code was:\n%s", buildErr, hookCode)
			}

			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(
				fmt.Sprintf("Hook %s appended to %s\n\n%s", methodName, modelPath, hookCode))), nil
		},
	}
}

// parseModelStructName finds the first exported non-"Service" struct in the file.
func parseModelStructName(path string) (string, error) {
	symbols, _, err := parseFileSymbols(filepath.Dir(path), path)
	if err != nil {
		return "", err
	}
	for _, s := range symbols {
		if s.Kind == "struct" && !strings.HasSuffix(s.Name, "Service") && !strings.HasSuffix(s.Name, "Repository") {
			return s.Name, nil
		}
	}
	return "", fmt.Errorf("no model struct found in %s", path)
}

// buildHookCode generates the GORM hook method.
func buildHookCode(modelStructName, methodName, description string) string {
	receiver := strings.ToLower(string([]rune(modelStructName)[0]))
	return fmt.Sprintf(
		`// %s is called by GORM %s a new %s.
// TODO: %s
func (%s *%s) %s(tx *gorm.DB) error {
	// Implement: %s
	return nil
}
`,
		methodName,
		hookEventDescription(methodName),
		modelStructName,
		description,
		receiver,
		modelStructName,
		methodName,
		description,
	)
}

// hookEventDescription returns a human-readable description of when the hook fires.
func hookEventDescription(methodName string) string {
	switch methodName {
	case "BeforeCreate":
		return "before inserting"
	case "AfterCreate":
		return "after inserting"
	case "BeforeUpdate":
		return "before updating"
	case "AfterUpdate":
		return "after updating"
	case "BeforeDelete":
		return "before deleting"
	case "AfterDelete":
		return "after deleting"
	default:
		return "on"
	}
}

// ---------------------------------------------------------------------------
// Plugin registration
// ---------------------------------------------------------------------------

func init() {
	RegisterPlugin(func(executor CommandExecutor, workingDir string) []RegisteredTool {
		return []RegisteredTool{
			generateMethodTool(executor, workingDir),
			addHookTool(executor, workingDir),
		}
	})
}
