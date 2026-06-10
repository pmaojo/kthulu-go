package mcp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

type GitService struct{}

type NoArgs struct{}
type LogArgs struct {
	Count int `json:"count" jsonschema:"description=Number of entries to show,default=10"`
}

func NewGitService() *GitService {
	return &GitService{}
}

func (s *GitService) GetTools(workingDir string) []RegisteredTool {
	return []RegisteredTool{
		s.statusTool(workingDir),
		s.diffTool(workingDir),
		s.logTool(workingDir),
	}
}

func (s *GitService) statusTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "git_status",
		Description: "Get the current git status of the project",
		Handler: func(ctx context.Context, _ NoArgs) (*mcp_golang.ToolResponse, error) {
			return s.runGit(ctx, resolveWorkdir(workingDir), "status")
		},
	}
}

func (s *GitService) diffTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "git_diff",
		Description: "Get the git diff of the current working tree",
		Handler: func(ctx context.Context, _ NoArgs) (*mcp_golang.ToolResponse, error) {
			return s.runGit(ctx, resolveWorkdir(workingDir), "diff")
		},
	}
}

func (s *GitService) logTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "git_log",
		Description: "Get the recent git log entries",
		Handler: func(ctx context.Context, args LogArgs) (*mcp_golang.ToolResponse, error) {
			count := args.Count
			if count == 0 {
				count = 10
			}
			return s.runGit(ctx, resolveWorkdir(workingDir), "log", fmt.Sprintf("-%d", count), "--oneline")
		},
	}
}

func (s *GitService) runGit(ctx context.Context, dir string, args ...string) (*mcp_golang.ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()

	output := string(out)
	if err != nil {
		output = fmt.Sprintf("Error running git %s: %v\nOutput: %s", strings.Join(args, " "), err, output)
	}

	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(output)), nil
}
