package tools

import (
	"context"
	"fmt"
	"strings"
)

// MCPToolWrapper wraps MCP tools to work with our tool system
type MCPToolWrapper struct {
	ServerName  string
	MCPTool     MCPToolInfo
	CallFunc    func(name string, args map[string]interface{}) (string, error)
}

// MCPToolInfo contains info about an MCP tool
type MCPToolInfo struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// ToNativeTool converts an MCP tool to our native Tool format
func (w *MCPToolWrapper) ToNativeTool() *Tool {
	return &Tool{
		Name:        fmt.Sprintf("mcp_%s_%s", w.ServerName, w.MCPTool.Name),
		Description: fmt.Sprintf("[MCP: %s] %s", w.ServerName, w.MCPTool.Description),
		Parameters:  w.MCPTool.InputSchema,
		NeedsApproval: true, // MCP tools should require approval
		Execute: func(ctx context.Context, args map[string]any) (*Result, error) {
			// Convert args to map[string]interface{}
			argsInterface := make(map[string]interface{})
			for k, v := range args {
				argsInterface[k] = v
			}

			output, err := w.CallFunc(w.MCPTool.Name, argsInterface)
			if err != nil {
				return &Result{
					Success: false,
					Output:  err.Error(),
					Display: formatMCPError(w.ServerName, w.MCPTool.Name, err),
					Error:   err.Error(),
				}, nil
			}

			return &Result{
				Success: true,
				Output:  output,
				Display: formatMCPOutput(w.ServerName, w.MCPTool.Name, output),
			}, nil
		},
	}
}

func formatMCPOutput(server, tool, output string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("╭─ MCP: %s/%s ", server, tool))
	sb.WriteString(strings.Repeat("─", 30))
	sb.WriteString("╮\n")

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line != "" {
			sb.WriteString(fmt.Sprintf("│ %s\n", line))
		}
	}

	sb.WriteString("╰" + strings.Repeat("─", 50) + " ✓ Done ─╯\n")
	return sb.String()
}

func formatMCPError(server, tool string, err error) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("╭─ MCP: %s/%s ", server, tool))
	sb.WriteString(strings.Repeat("─", 30))
	sb.WriteString("╮\n")
	sb.WriteString(fmt.Sprintf("│ ❌ Error: %v\n", err))
	sb.WriteString("╰" + strings.Repeat("─", 50) + " ✗ Failed ─╯\n")

	return sb.String()
}

// RegisterMCPTools registers all tools from MCP servers into the registry
func RegisterMCPTools(registry *Registry, tools []MCPToolWrapper) {
	for _, wrapper := range tools {
		registry.Register(wrapper.ToNativeTool())
	}
}
