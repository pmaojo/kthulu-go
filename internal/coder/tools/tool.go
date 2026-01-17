package tools

import (
	"context"
	"encoding/json"
)

// Tool represents an AI-callable tool
type Tool struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Parameters    map[string]interface{} `json:"parameters"` // JSON Schema
	NeedsApproval bool                   `json:"-"`
	Execute       ExecuteFunc            `json:"-"`
}

// ExecuteFunc is the function signature for tool execution
type ExecuteFunc func(ctx context.Context, args map[string]any) (*Result, error)

// Result represents the output of a tool execution
type Result struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`  // Text returned to AI
	Display string `json:"display"` // Rich text for TUI (can include ANSI)
	Error   string `json:"error,omitempty"`
}

// ToolCall represents a request from the AI to use a tool
type ToolCall struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Args     map[string]any `json:"arguments"`
	Approved bool           `json:"-"`
}

// Registry holds all available tools
type Registry struct {
	tools map[string]*Tool
}

// NewRegistry creates a new tool registry with default tools
func NewRegistry() *Registry {
	r := &Registry{
		tools: make(map[string]*Tool),
	}

	// Register default tools
	r.Register(BashTool)
	r.Register(ReadFileTool)
	r.Register(WriteFileTool)
	r.Register(GrepTool)
	r.Register(ThinkTool)
	r.Register(KthuluTool)
	r.Register(AnalysisTool)

	return r
}

// Register adds a tool to the registry
func (r *Registry) Register(tool *Tool) {
	r.tools[tool.Name] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (*Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// All returns all registered tools
func (r *Registry) All() []*Tool {
	tools := make([]*Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// ToOpenAIFormat converts tools to OpenAI function calling format
func (r *Registry) ToOpenAIFormat() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		})
	}
	return result
}

// ParseToolCall parses a tool call from AI response
func ParseToolCall(data []byte) (*ToolCall, error) {
	var call ToolCall
	if err := json.Unmarshal(data, &call); err != nil {
		return nil, err
	}
	return &call, nil
}
