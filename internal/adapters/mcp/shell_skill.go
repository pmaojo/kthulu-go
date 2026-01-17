package mcp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

type ShellService struct{}

type ShellArgs struct {
	Command string `json:"command" jsonschema:"description=The command to execute,required"`
}

func NewShellService() *ShellService {
	return &ShellService{}
}

func (s *ShellService) ExecuteTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "host_shell_execute",
		Description: "Execute a command in the host shell. Use this for running tests or system ops. BLOCKED: 'rm -rf /'",
		Handler: func(ctx context.Context, args ShellArgs) (*mcp_golang.ToolResponse, error) {
			cmdStr := args.Command
			if cmdStr == "" {
				return nil, fmt.Errorf("command argument required")
			}

			// Security: Basic blacklist
			if strings.Contains(cmdStr, "rm -rf /") {
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("❌ Operation blocked by safety policy")), nil
			}

			cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
			cmd.Dir = workingDir
			out, err := cmd.CombinedOutput()

			output := string(out)
			if err != nil {
				output = fmt.Sprintf("%s\nError: %v", output, err)
			}

			// Truncate if too long (50KB limit)
			if len(output) > 50000 {
				output = output[:50000] + "\n...[Output Truncated]..."
			}

			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(output)), nil
		},
	}
}
