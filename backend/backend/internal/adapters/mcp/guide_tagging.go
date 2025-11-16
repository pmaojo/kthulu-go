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

// GuideTaggingService builds guidance for applying @kthulu tags across a project.
type GuideTaggingService struct {
	parser *parser.TagParser
}

// TaggingGuideArguments optionally scopes the guide to a sub directory.
type TaggingGuideArguments struct {
	Focus string `json:"focus,omitempty" jsonschema:"description=Optional relative path inside the working directory to limit the analysis"`
}

// tagCoverage contains the per-file analysis used to build the guide.
type tagCoverage struct {
	FileAnalyses map[string]*parser.FileAnalysis
	Untagged     []string
}

// NewGuideTaggingService constructs the service.
func NewGuideTaggingService(p *parser.TagParser) *GuideTaggingService {
	return &GuideTaggingService{parser: p}
}

// Tool returns the MCP tool registration for the guide_tagging tool.
func (s *GuideTaggingService) Tool(projectPath string) RegisteredTool {
	handler := func(ctx context.Context, args TaggingGuideArguments) (*mcp_golang.ToolResponse, error) {
		target := projectPath
		if args.Focus != "" {
			target = filepath.Join(projectPath, args.Focus)
		}

		guide, err := s.BuildGuide(target)
		if err != nil {
			return nil, err
		}

		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(guide)), nil
	}

	description := "Analyze the current working directory and explain how to add @kthulu tags to untagged Go files."
	return RegisteredTool{Name: "guide_tagging", Description: description, Handler: handler}
}

// BuildGuide creates a textual guide describing missing tags.
func (s *GuideTaggingService) BuildGuide(projectPath string) (string, error) {
	info, err := os.Stat(projectPath)
	if err != nil {
		return "", fmt.Errorf("unable to access %s: %w", projectPath, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", projectPath)
	}

	analysis, err := s.parser.AnalyzeProject(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to analyze project: %w", err)
	}

	coverage, err := s.collectCoverage(projectPath)
	if err != nil {
		return "", err
	}

	rel := func(path string) string {
		relative, err := filepath.Rel(projectPath, path)
		if err != nil {
			return path
		}
		return relative
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Kthulu tagging guide for %s\n", projectPath))
	builder.WriteString(fmt.Sprintf("Detected %d modules and %d existing tags.\n", len(analysis.Modules), len(analysis.Tags)))

	if len(coverage.Untagged) == 0 {
		builder.WriteString("\nAll Go files already include @kthulu annotations. Great job!\n")
		return builder.String(), nil
	}

	builder.WriteString("\nFocus on the following files to improve tagging coverage:\n")

	for _, file := range coverage.Untagged {
		fileAnalysis := coverage.FileAnalyses[file]
		builder.WriteString(fmt.Sprintf("\n• %s (package %s)\n", rel(file), fileAnalysis.Package))

		inferredModule := inferModuleCandidate(fileAnalysis, analysis.Modules)
		if inferredModule != "" {
			builder.WriteString(fmt.Sprintf("  - Add @kthulu:module:%s near the top of the file.\n", inferredModule))
		} else {
			builder.WriteString("  - Add @kthulu:module:<module-name> to register this file.\n")
		}

		hints := symbolHints(fileAnalysis.Symbols)
		if len(hints) > 0 {
			builder.WriteString("  - Suggested semantic tags:\n")
			for _, hint := range hints {
				builder.WriteString(fmt.Sprintf("    • %s\n", hint))
			}
		} else {
			builder.WriteString("  - Exported types or functions were not detected; add @kthulu:service or @kthulu:handler tags where appropriate.\n")
		}
	}

	builder.WriteString("\nAfter tagging, rerun this tool to verify that coverage improved.\n")

	return builder.String(), nil
}

func (s *GuideTaggingService) collectCoverage(projectPath string) (*tagCoverage, error) {
	coverage := &tagCoverage{FileAnalyses: make(map[string]*parser.FileAnalysis)}

	err := filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if shouldIgnoreDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fileAnalysis, err := s.parser.AnalyzeFile(path)
		if err != nil {
			return fmt.Errorf("failed to analyze %s: %w", path, err)
		}
		coverage.FileAnalyses[path] = fileAnalysis
		if len(fileAnalysis.Tags) == 0 {
			coverage.Untagged = append(coverage.Untagged, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(coverage.Untagged)
	return coverage, nil
}

func shouldIgnoreDir(name string) bool {
	switch name {
	case ".git", "vendor", "node_modules":
		return true
	default:
		return false
	}
}

func inferModuleCandidate(file *parser.FileAnalysis, modules map[string]*parser.Module) string {
	if file == nil {
		return ""
	}

	pkg := strings.ToLower(file.Package)
	for moduleName := range modules {
		if strings.EqualFold(moduleName, pkg) {
			return moduleName
		}
	}

	dirName := strings.ToLower(filepath.Base(filepath.Dir(file.FilePath)))
	if dirName != "" {
		return dirName
	}

	return ""
}

func symbolHints(symbols []parser.Symbol) []string {
	suggestions := make([]string, 0)
	seen := make(map[string]struct{})
	for _, sym := range symbols {
		if hint := classifySymbol(sym); hint != "" {
			if _, ok := seen[hint]; ok {
				continue
			}
			seen[hint] = struct{}{}
			suggestions = append(suggestions, hint)
		}
	}
	sort.Strings(suggestions)
	return suggestions
}

func classifySymbol(symbol parser.Symbol) string {
	lower := strings.ToLower(symbol.Name)
	switch {
	case strings.Contains(lower, "handler"):
		return fmt.Sprintf("@kthulu:handler:%s", symbol.Name)
	case strings.Contains(lower, "service"):
		return fmt.Sprintf("@kthulu:service:%s", symbol.Name)
	case strings.Contains(lower, "repository") || strings.Contains(lower, "repo"):
		return fmt.Sprintf("@kthulu:repository:%s", symbol.Name)
	case symbol.Type == "type":
		return fmt.Sprintf("@kthulu:domain:%s", symbol.Name)
	default:
		return ""
	}
}
