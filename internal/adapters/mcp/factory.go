package mcp

import (
	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
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

	bdd := NewBDDService()
	tools = append(tools,
		bdd.ListFeaturesTool(workingDir),
		bdd.ReadFeatureTool(workingDir),
		bdd.RunScenarioTool(workingDir),
	)

	// Native Skills (Batteries Included)
	shell := NewShellService()
	tools = append(tools, shell.ExecuteTool(workingDir))

	git := NewGitService()
	tools = append(tools, git.GetTools(workingDir)...)

	fs := NewFileSystemService()
	tools = append(tools, fs.GetTools(workingDir)...)

	search := NewSearchService()
	tools = append(tools, search.GetTools(workingDir)...)

	goAST := NewGoASTService()
	tools = append(tools, goAST.GetTools(workingDir)...)

	database := NewDatabaseService()
	tools = append(tools, database.GetTools(workingDir)...)

	goTest := NewGoTestService()
	tools = append(tools, goTest.GetTools(workingDir)...)

	watch := NewFileWatchService()
	tools = append(tools, watch.GetTools(workingDir)...)

	return tools
}
