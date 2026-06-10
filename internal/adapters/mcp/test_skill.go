package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// GoTestService exposes native Go toolchain commands (test, build, vet) as
// structured MCP tools, complementing the BDD-only test runner.
type GoTestService struct{}

// NewGoTestService creates a new GoTestService.
func NewGoTestService() *GoTestService {
	return &GoTestService{}
}

// GetTools returns all Go toolchain tools bound to the working directory.
func (s *GoTestService) GetTools(workingDir string) []RegisteredTool {
	return []RegisteredTool{
		s.testTool(workingDir),
		s.buildTool(workingDir),
		s.vetTool(workingDir),
	}
}

// GoTestArgs defines arguments for running Go tests.
type GoTestArgs struct {
	Packages string `json:"packages,omitempty" jsonschema:"description=Package pattern to test (default ./...)"`
	Run      string `json:"run,omitempty" jsonschema:"description=Regex passed to -run to select tests"`
	Verbose  bool   `json:"verbose,omitempty" jsonschema:"description=Enable -v output"`
	Race     bool   `json:"race,omitempty" jsonschema:"description=Enable the race detector"`
	Cover    bool   `json:"cover,omitempty" jsonschema:"description=Report code coverage"`
	Timeout  string `json:"timeout,omitempty" jsonschema:"description=Test timeout such as 60s or 5m (default 5m)"`
	Count    int    `json:"count,omitempty" jsonschema:"description=Run each test N times; 1 disables test caching"`
}

// GoPackagesArgs defines arguments for build and vet.
type GoPackagesArgs struct {
	Packages string `json:"packages,omitempty" jsonschema:"description=Package pattern (default ./...)"`
}

func (s *GoTestService) testTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "go_test",
		Description: "Run Go tests natively with go test. Supports -run filters, race detector, coverage, and timeouts. Failures are returned as output rather than errors so results stay inspectable.",
		Handler: func(ctx context.Context, args GoTestArgs) (*mcp_golang.ToolResponse, error) {
			result, err := s.RunTests(ctx, resolveWorkdir(workingDir), args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}

func (s *GoTestService) buildTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "go_build",
		Description: "Compile Go packages with go build and report any compilation errors.",
		Handler: func(ctx context.Context, args GoPackagesArgs) (*mcp_golang.ToolResponse, error) {
			result, err := s.runGo(ctx, resolveWorkdir(workingDir), "build", normalizePackages(args.Packages))
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}

func (s *GoTestService) vetTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "go_vet",
		Description: "Run go vet static analysis on Go packages and report findings.",
		Handler: func(ctx context.Context, args GoPackagesArgs) (*mcp_golang.ToolResponse, error) {
			result, err := s.runGo(ctx, resolveWorkdir(workingDir), "vet", normalizePackages(args.Packages))
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}

// RunTests executes go test with the requested options.
func (s *GoTestService) RunTests(ctx context.Context, workingDir string, args GoTestArgs) (string, error) {
	cmdArgs := []string{"test"}

	if args.Verbose {
		cmdArgs = append(cmdArgs, "-v")
	}
	if args.Race {
		cmdArgs = append(cmdArgs, "-race")
	}
	if args.Cover {
		cmdArgs = append(cmdArgs, "-cover")
	}
	if args.Run != "" {
		cmdArgs = append(cmdArgs, "-run", args.Run)
	}
	if args.Count > 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("-count=%d", args.Count))
	}
	timeout := strings.TrimSpace(args.Timeout)
	if timeout == "" {
		timeout = "5m"
	}
	cmdArgs = append(cmdArgs, "-timeout", timeout, normalizePackages(args.Packages))

	return s.runGo(ctx, workingDir, cmdArgs[0], cmdArgs[1:]...)
}

func (s *GoTestService) runGo(ctx context.Context, workingDir, subcommand string, args ...string) (string, error) {
	cmdArgs := append([]string{subcommand}, args...)
	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Dir = workingDir
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	label := "go " + strings.Join(cmdArgs, " ")
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("$ %s\n", label))
	if output != "" {
		builder.WriteString(output)
		builder.WriteString("\n")
	}
	if err != nil {
		builder.WriteString(fmt.Sprintf("\nExit status: %v", err))
	} else {
		builder.WriteString("\nExit status: ok")
	}

	return truncateOutput(builder.String()), nil
}

func normalizePackages(packages string) string {
	trimmed := strings.TrimSpace(packages)
	if trimmed == "" {
		return "./..."
	}
	return trimmed
}
