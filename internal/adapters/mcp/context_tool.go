package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// contextFileName is stored alongside the project (or in the session workdir).
const contextFileName = ".kthulu-context.json"

type projectContext struct {
	Notes     map[string]string `json:"notes"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func loadContext(dir string) projectContext {
	data, err := os.ReadFile(filepath.Join(dir, contextFileName))
	if err != nil {
		return projectContext{Notes: map[string]string{}}
	}
	var c projectContext
	if err := json.Unmarshal(data, &c); err != nil {
		return projectContext{Notes: map[string]string{}}
	}
	if c.Notes == nil {
		c.Notes = map[string]string{}
	}
	return c
}

func saveContext(dir string, c projectContext) error {
	c.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, contextFileName), data, 0o644)
}

// ContextSetArgs for context_set tool.
type ContextSetArgs struct {
	Key   string `json:"key" jsonschema:"required,description=Context key (e.g. 'currency', 'vat_rate', 'business_rule')."`
	Value string `json:"value" jsonschema:"required,description=Value to store. Use 'DELETE' to remove the key."`
}

// ContextGetArgs for context_get tool.
type ContextGetArgs struct {
	Key string `json:"key,omitempty" jsonschema:"description=Specific key to retrieve. Omit to return all stored context."`
}

func contextTools(workingDir string) []RegisteredTool {
	dir := resolveWorkdir(workingDir)

	set := RegisteredTool{
		Name: "context_set",
		Description: "Store a persistent business-context note for this project (survives across sessions). " +
			"Use it to record domain rules, conventions, and decisions that should inform all future code generation. " +
			"Examples: context_set currency EUR — context_set vat_rate 0.21 — context_set prices_in_cents true. " +
			"Pass value='DELETE' to remove a key.",
		Handler: func(ctx context.Context, args ContextSetArgs) (*mcp_golang.ToolResponse, error) {
			if strings.TrimSpace(args.Key) == "" {
				return nil, fmt.Errorf("key is required")
			}
			c := loadContext(dir)
			if args.Value == "DELETE" {
				delete(c.Notes, args.Key)
				if err := saveContext(dir, c); err != nil {
					return nil, err
				}
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(
					fmt.Sprintf("🗑️  Removed context key %q", args.Key))), nil
			}
			c.Notes[args.Key] = args.Value
			if err := saveContext(dir, c); err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(
				fmt.Sprintf("✅ context[%q] = %q  (stored in %s)", args.Key, args.Value, contextFileName))), nil
		},
	}

	get := RegisteredTool{
		Name: "context_get",
		Description: "Retrieve stored business-context notes for this project. " +
			"Call with no key to dump all notes. Useful at the start of any session to recall project conventions.",
		Handler: func(ctx context.Context, args ContextGetArgs) (*mcp_golang.ToolResponse, error) {
			c := loadContext(dir)
			if args.Key != "" {
				v, ok := c.Notes[args.Key]
				if !ok {
					return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(
						fmt.Sprintf("(no context stored for key %q)", args.Key))), nil
				}
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(
					fmt.Sprintf("%s = %s", args.Key, v))), nil
			}
			if len(c.Notes) == 0 {
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(
					"No project context stored yet. Use context_set to record business rules and conventions.")), nil
			}
			var lines []string
			for k, v := range c.Notes {
				lines = append(lines, fmt.Sprintf("  %-24s %s", k, v))
			}
			out := fmt.Sprintf("📋 Project context (last updated %s):\n%s",
				c.UpdatedAt.Format("2006-01-02 15:04"), strings.Join(lines, "\n"))
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(out)), nil
		},
	}

	return []RegisteredTool{set, get}
}

func init() {
	RegisterPlugin(func(_ CommandExecutor, workingDir string) []RegisteredTool {
		return contextTools(workingDir)
	})
}
