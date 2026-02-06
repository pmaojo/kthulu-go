package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// BDDService provides tools for managing BDD features and scenarios.
type BDDService struct{}

// NewBDDService creates a new BDDService.
func NewBDDService() *BDDService {
	return &BDDService{}
}

// ListFeaturesArgs defines arguments for listing features (none required).
type ListFeaturesArgs struct{}

// ListFeaturesTool returns a tool that lists all .feature files in the project.
func (s *BDDService) ListFeaturesTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "list_features",
		Description: "List all Gherkin .feature files in the project",
		Handler: func(ctx context.Context, args ListFeaturesArgs) (*mcp_golang.ToolResponse, error) {
			var features []string

			// Default search paths
			searchPaths := []string{"features", "backend/features"}

			found := false
			for _, searchPath := range searchPaths {
				fullPath := filepath.Join(workingDir, searchPath)
				if _, err := os.Stat(fullPath); err == nil {
					found = true
					err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if !info.IsDir() && strings.HasSuffix(info.Name(), ".feature") {
							relPath, _ := filepath.Rel(workingDir, path)
							features = append(features, relPath)
						}
						return nil
					})
					if err != nil {
						return nil, fmt.Errorf("failed to walk features dir: %w", err)
					}
				}
			}

			if !found {
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("No features directory found (checked ./features and ./backend/features)")), nil
			}

			if len(features) == 0 {
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("No .feature files found.")), nil
			}

			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(strings.Join(features, "\n"))), nil
		},
	}
}

// ReadFeatureArgs defines arguments for reading a feature file.
type ReadFeatureArgs struct {
	Path string `json:"path" jsonschema:"description=Path to the .feature file"`
}

// ReadFeatureTool returns a tool that reads the content of a specific .feature file.
func (s *BDDService) ReadFeatureTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "read_feature",
		Description: "Read the content of a Gherkin .feature file",
		Handler: func(ctx context.Context, args ReadFeatureArgs) (*mcp_golang.ToolResponse, error) {
			path := args.Path
			if path == "" {
				return nil, fmt.Errorf("argument 'path' is required")
			}

			fullPath := filepath.Join(workingDir, path)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %w", path, err)
			}

			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(content))), nil
		},
	}
}

// RunScenarioArgs defines arguments for running scenarios.
type RunScenarioArgs struct {
	Filter string `json:"filter,omitempty" jsonschema:"description=Regex filter for scenarios or feature name"`
}

// RunScenarioTool returns a tool that executes a specific BDD scenario or all scenarios.
func (s *BDDService) RunScenarioTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "run_scenario",
		Description: "Run BDD scenarios using godog (via go test)",
		Handler: func(ctx context.Context, args RunScenarioArgs) (*mcp_golang.ToolResponse, error) {
			// Arg1 can be a regex or name filter
			filter := args.Filter

			// Determine where the tests are.
			testPath := "./..."
			if _, err := os.Stat(filepath.Join(workingDir, "backend/features")); err == nil {
				testPath = "./backend/features/..."
			} else if _, err := os.Stat(filepath.Join(workingDir, "features")); err == nil {
				testPath = "./features/..."
			}

			cmdArgs := []string{"test", "-v", testPath}

			if filter != "" {
				cmdArgs = append(cmdArgs, "-args", filter)
			}

			cmd := exec.CommandContext(ctx, "go", cmdArgs...)
			cmd.Dir = workingDir
			cmd.Env = os.Environ()

			output, err := cmd.CombinedOutput()

			result := string(output)
			if err != nil {
				return nil, fmt.Errorf("tests failed:\n%s\nError: %w", result, err)
			}

			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}
