package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"

	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/parser"
)

// ProjectInsightsService exposes parser-driven project analysis as MCP tools.
type ProjectInsightsService struct {
	parser *parser.TagParser
}

type projectInsightArgs struct{}

// NewProjectInsightsService creates a new insights service.
func NewProjectInsightsService(p *parser.TagParser) *ProjectInsightsService {
	return &ProjectInsightsService{parser: p}
}

// Tool registrations -------------------------------------------------------

// OverviewTool returns a tool that summarizes the project.
func (s *ProjectInsightsService) OverviewTool(projectPath string) RegisteredTool {
	handler := func(ctx context.Context, _ projectInsightArgs) (*mcp_golang.ToolResponse, error) {
		overview, err := s.BuildOverview(projectPath)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(overview)), nil
	}

	description := "Summarize the project: module, dependency, and tag counts with highlights."
	return RegisteredTool{Name: "project_overview", Description: description, Handler: handler}
}

// ModulesTool returns a tool that lists modules, packages, files, and dependencies.
func (s *ProjectInsightsService) ModulesTool(projectPath string) RegisteredTool {
	handler := func(ctx context.Context, _ projectInsightArgs) (*mcp_golang.ToolResponse, error) {
		description, err := s.DescribeModules(projectPath)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(description)), nil
	}

	description := "Detail each @kthulu module with its package, files, and declared dependencies."
	return RegisteredTool{Name: "project_modules", Description: description, Handler: handler}
}

// TagsTool returns a tool that summarizes tag types across the project.
func (s *ProjectInsightsService) TagsTool(projectPath string) RegisteredTool {
	handler := func(ctx context.Context, _ projectInsightArgs) (*mcp_golang.ToolResponse, error) {
		description, err := s.DescribeTags(projectPath)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(description)), nil
	}

	description := "Summarize how many of each @kthulu tag type were detected."
	return RegisteredTool{Name: "project_tags", Description: description, Handler: handler}
}

// DependenciesTool returns a tool that lists module dependency edges.
func (s *ProjectInsightsService) DependenciesTool(projectPath string) RegisteredTool {
	handler := func(ctx context.Context, _ projectInsightArgs) (*mcp_golang.ToolResponse, error) {
		description, err := s.DescribeDependencies(projectPath)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(description)), nil
	}

	description := "List the declared module dependencies detected during parsing."
	return RegisteredTool{Name: "project_dependencies", Description: description, Handler: handler}
}

// Builders ----------------------------------------------------------------

// BuildOverview summarizes the project analysis in text form.
func (s *ProjectInsightsService) BuildOverview(projectPath string) (string, error) {
	analysis, err := s.analyze(projectPath)
	if err != nil {
		return "", err
	}

	moduleNames := sortedKeys(analysis.Modules)

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Kthulu project overview for %s\n", projectPath))
	builder.WriteString(fmt.Sprintf("Modules: %d\n", len(moduleNames)))
	builder.WriteString(fmt.Sprintf("Dependencies: %d\n", len(analysis.Dependencies)))
	builder.WriteString(fmt.Sprintf("Tags: %d\n", len(analysis.Tags)))

	if len(moduleNames) > 0 {
		builder.WriteString("\nModules:\n")
		for _, name := range moduleNames {
			module := analysis.Modules[name]
			builder.WriteString(fmt.Sprintf("• %s (package %s, %d files)\n", name, module.Package, len(module.Files)))
		}
	} else {
		builder.WriteString("\nNo @kthulu:module annotations were found. Use guide_tagging to add them.\n")
	}

	if len(analysis.Dependencies) > 0 {
		builder.WriteString("\nDependency edges detected:\n")
		for _, dep := range sortDependencies(analysis.Dependencies) {
			builder.WriteString(fmt.Sprintf("• %s -> %s (%s)\n", dep.From, dep.To, dep.Type))
		}
	}

	return builder.String(), nil
}

// DescribeModules enumerates modules with metadata.
func (s *ProjectInsightsService) DescribeModules(projectPath string) (string, error) {
	analysis, err := s.analyze(projectPath)
	if err != nil {
		return "", err
	}

	moduleNames := sortedKeys(analysis.Modules)

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Modules detected in %s\n", projectPath))
	if len(moduleNames) == 0 {
		builder.WriteString("\nNo modules detected. Add @kthulu:module tags to your Go files.\n")
		return builder.String(), nil
	}

	for _, name := range moduleNames {
		module := analysis.Modules[name]
		relFiles := relativeFileList(projectPath, module.Files)
		deps := append([]string{}, module.Dependencies...)
		sort.Strings(deps)

		builder.WriteString(fmt.Sprintf("\n• %s\n", name))
		builder.WriteString(fmt.Sprintf("  Package: %s\n", module.Package))
		builder.WriteString(fmt.Sprintf("  Files: %d\n", len(relFiles)))
		for _, file := range relFiles {
			builder.WriteString(fmt.Sprintf("    - %s\n", file))
		}
		if len(deps) > 0 {
			builder.WriteString("  Dependencies:\n")
			for _, dep := range deps {
				builder.WriteString(fmt.Sprintf("    - %s\n", dep))
			}
		} else {
			builder.WriteString("  Dependencies: <none>\n")
		}
	}

	return builder.String(), nil
}

// DescribeTags groups tags by type with counts.
func (s *ProjectInsightsService) DescribeTags(projectPath string) (string, error) {
	analysis, err := s.analyze(projectPath)
	if err != nil {
		return "", err
	}

	counts := make(map[parser.TagType]int)
	for _, tag := range analysis.Tags {
		counts[tag.Type]++
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Tag summary for %s\n", projectPath))
	if len(counts) == 0 {
		builder.WriteString("\nNo @kthulu tags were detected.\n")
		return builder.String(), nil
	}

	types := make([]string, 0, len(counts))
	for tagType := range counts {
		types = append(types, string(tagType))
	}
	sort.Strings(types)

	builder.WriteString("\nTag counts by type:\n")
	for _, tagType := range types {
		builder.WriteString(fmt.Sprintf("• %s: %d\n", tagType, counts[parser.TagType(tagType)]))
	}

	return builder.String(), nil
}

// DescribeDependencies lists dependency edges.
func (s *ProjectInsightsService) DescribeDependencies(projectPath string) (string, error) {
	analysis, err := s.analyze(projectPath)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Dependencies detected in %s\n", projectPath))

	if len(analysis.Dependencies) == 0 {
		builder.WriteString("\nNo @kthulu:dependency tags were found.\n")
		return builder.String(), nil
	}

	builder.WriteString("\n")
	for _, dep := range sortDependencies(analysis.Dependencies) {
		builder.WriteString(fmt.Sprintf("• %s -> %s (%s)\n", dep.From, dep.To, dep.Type))
	}

	return builder.String(), nil
}

// Helpers -----------------------------------------------------------------

func (s *ProjectInsightsService) analyze(projectPath string) (*parser.ProjectAnalysis, error) {
	info, err := os.Stat(projectPath)
	if err != nil {
		return nil, fmt.Errorf("unable to access %s: %w", projectPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", projectPath)
	}

	analysis, err := s.parser.AnalyzeProject(projectPath)
	if err != nil {
		return nil, fmt.Errorf("analyze project: %w", err)
	}
	return analysis, nil
}

func sortedKeys(modules map[string]*parser.Module) []string {
	names := make([]string, 0, len(modules))
	for name := range modules {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func relativeFileList(base string, files []string) []string {
	rel := make([]string, 0, len(files))
	for _, file := range files {
		relative, err := filepath.Rel(base, file)
		if err != nil {
			relative = file
		}
		rel = append(rel, relative)
	}
	sort.Strings(rel)
	return rel
}

func sortDependencies(deps []parser.Dependency) []parser.Dependency {
	sorted := append([]parser.Dependency(nil), deps...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].From == sorted[j].From {
			if sorted[i].To == sorted[j].To {
				return sorted[i].Type < sorted[j].Type
			}
			return sorted[i].To < sorted[j].To
		}
		return sorted[i].From < sorted[j].From
	})
	return sorted
}
