package mcpserver

import (
	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/parser"
	"github.com/spf13/cobra"
)

// ToolFactory assembles all MCP tools that the CLI exposes.
type ToolFactory struct {
	root     *cobra.Command
	executor CommandExecutor
	parser   *parser.TagParser
}

// NewToolFactory constructs a new ToolFactory.
func NewToolFactory(root *cobra.Command, executor CommandExecutor, tagParser *parser.TagParser) *ToolFactory {
	return &ToolFactory{root: root, executor: executor, parser: tagParser}
}

// BuildTools returns the list of registered tools for the given working directory.
func (f *ToolFactory) BuildTools(workingDir string, filter CommandFilter) []RegisteredTool {
	tools := BuildCommandTools(f.root, f.executor, workingDir, filter)

	guide := NewGuideTaggingService(f.parser)
	tools = append(tools, guide.Tool(workingDir))

	insights := NewProjectInsightsService(f.parser)
	tools = append(tools,
		insights.OverviewTool(workingDir),
		insights.ModulesTool(workingDir),
		insights.TagsTool(workingDir),
		insights.DependenciesTool(workingDir),
	)

	return tools
}
