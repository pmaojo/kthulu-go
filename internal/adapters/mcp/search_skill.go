package mcp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// SearchService exposes fast project-wide code search tools: regex content
// search with context lines and glob-based file discovery.
type SearchService struct{}

// NewSearchService creates a new SearchService.
func NewSearchService() *SearchService {
	return &SearchService{}
}

// GetTools returns all search tools bound to the working directory.
func (s *SearchService) GetTools(workingDir string) []RegisteredTool {
	return []RegisteredTool{
		s.codeSearchTool(workingDir),
		s.fileGlobTool(workingDir),
	}
}

// CodeSearchArgs defines arguments for regex content search.
type CodeSearchArgs struct {
	Pattern         string `json:"pattern" jsonschema:"description=Regular expression to search for (Go regexp syntax),required"`
	Path            string `json:"path,omitempty" jsonschema:"description=File or directory to search (default: project root)"`
	Include         string `json:"include,omitempty" jsonschema:"description=Glob filter on file names e.g. *.go or *.{ts,tsx}"`
	CaseInsensitive bool   `json:"case_insensitive,omitempty" jsonschema:"description=Match case-insensitively"`
	Context         int    `json:"context,omitempty" jsonschema:"description=Lines of context to show around each match"`
	MaxResults      int    `json:"max_results,omitempty" jsonschema:"description=Maximum number of matches to return (default 100)"`
}

// FileGlobArgs defines arguments for glob-based file discovery.
type FileGlobArgs struct {
	Pattern    string `json:"pattern" jsonschema:"description=Glob pattern such as **/*.go or internal/**/handler_*.go,required"`
	Path       string `json:"path,omitempty" jsonschema:"description=Directory to search (default: project root)"`
	MaxResults int    `json:"max_results,omitempty" jsonschema:"description=Maximum number of paths to return (default 200)"`
}

func (s *SearchService) codeSearchTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "code_search",
		Description: "Search file contents with a regular expression. Returns path:line: text matches with optional context lines. Skips binary files and build/VCS directories.",
		Handler: func(ctx context.Context, args CodeSearchArgs) (*mcp_golang.ToolResponse, error) {
			result, err := s.CodeSearch(resolveWorkdir(workingDir), args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}

func (s *SearchService) fileGlobTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "file_glob",
		Description: "Find files by glob pattern. Supports ** for recursive matching, e.g. **/*_test.go.",
		Handler: func(ctx context.Context, args FileGlobArgs) (*mcp_golang.ToolResponse, error) {
			result, err := s.FileGlob(resolveWorkdir(workingDir), args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}

// CodeSearch runs a regex search across project files.
func (s *SearchService) CodeSearch(workingDir string, args CodeSearchArgs) (string, error) {
	if strings.TrimSpace(args.Pattern) == "" {
		return "", fmt.Errorf("argument 'pattern' is required")
	}

	pattern := args.Pattern
	if args.CaseInsensitive {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid pattern: %w", err)
	}

	searchPath := args.Path
	if strings.TrimSpace(searchPath) == "" {
		searchPath = "."
	}
	root, err := resolveWorkspacePath(workingDir, searchPath)
	if err != nil {
		return "", err
	}

	maxResults := args.MaxResults
	if maxResults <= 0 {
		maxResults = 100
	}

	var builder strings.Builder
	matches := 0
	filesScanned := 0

	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			if shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if matches >= maxResults {
			return filepath.SkipAll
		}
		if args.Include != "" {
			if ok, _ := matchGlobName(args.Include, d.Name()); !ok {
				return nil
			}
		}

		rel, relErr := filepath.Rel(workingDir, path)
		if relErr != nil {
			rel = path
		}

		fileMatches, found, scanned := searchFile(path, rel, re, args.Context, maxResults-matches)
		if scanned {
			filesScanned++
		}
		if found > 0 {
			builder.WriteString(fileMatches)
			matches += found
		}
		return nil
	})
	if walkErr != nil {
		return "", fmt.Errorf("search failed: %w", walkErr)
	}

	if matches == 0 {
		return fmt.Sprintf("No matches for %q (%d files scanned)", args.Pattern, filesScanned), nil
	}

	header := fmt.Sprintf("%d match(es) for %q in %d scanned files:\n", matches, args.Pattern, filesScanned)
	return truncateOutput(header + builder.String()), nil
}

// searchFile scans a single file for regex matches and renders them with context.
func searchFile(path, displayPath string, re *regexp.Regexp, contextLines, remaining int) (string, int, bool) {
	data, err := os.ReadFile(path)
	if err != nil || len(data) > 2*1024*1024 || isBinary(data) {
		return "", 0, false
	}

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var builder strings.Builder
	found := 0
	for i, line := range lines {
		if found >= remaining {
			break
		}
		if !re.MatchString(line) {
			continue
		}
		found++

		start := i - contextLines
		if start < 0 {
			start = 0
		}
		end := i + contextLines
		if end >= len(lines) {
			end = len(lines) - 1
		}
		for j := start; j <= end; j++ {
			marker := "  "
			if j == i {
				marker = "\n→ "
			}
			builder.WriteString(fmt.Sprintf("%s%s:%d: %s\n", marker, displayPath, j+1, lines[j]))
		}
	}

	return builder.String(), found, true
}

// FileGlob finds files matching a glob pattern.
func (s *SearchService) FileGlob(workingDir string, args FileGlobArgs) (string, error) {
	if strings.TrimSpace(args.Pattern) == "" {
		return "", fmt.Errorf("argument 'pattern' is required")
	}

	searchPath := args.Path
	if strings.TrimSpace(searchPath) == "" {
		searchPath = "."
	}
	root, err := resolveWorkspacePath(workingDir, searchPath)
	if err != nil {
		return "", err
	}

	maxResults := args.MaxResults
	if maxResults <= 0 {
		maxResults = 200
	}

	re, err := globToRegexp(args.Pattern)
	if err != nil {
		return "", fmt.Errorf("invalid glob pattern: %w", err)
	}

	var results []string
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}
		rel = filepath.ToSlash(rel)
		if re.MatchString(rel) || re.MatchString(filepath.Base(rel)) {
			results = append(results, rel)
			if len(results) >= maxResults {
				return filepath.SkipAll
			}
		}
		return nil
	})
	if walkErr != nil {
		return "", fmt.Errorf("glob search failed: %w", walkErr)
	}

	if len(results) == 0 {
		return fmt.Sprintf("No files match %q", args.Pattern), nil
	}
	header := fmt.Sprintf("%d file(s) matching %q:\n", len(results), args.Pattern)
	return truncateOutput(header + strings.Join(results, "\n")), nil
}

// matchGlobName matches a file name against a glob that may use brace
// alternation, e.g. *.{go,md}.
func matchGlobName(pattern, name string) (bool, error) {
	for _, expanded := range expandBraces(pattern) {
		ok, err := filepath.Match(expanded, name)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// expandBraces expands a single {a,b,c} alternation group in a glob pattern.
func expandBraces(pattern string) []string {
	open := strings.Index(pattern, "{")
	close := strings.Index(pattern, "}")
	if open == -1 || close == -1 || close < open {
		return []string{pattern}
	}
	var expanded []string
	for _, alt := range strings.Split(pattern[open+1:close], ",") {
		expanded = append(expanded, expandBraces(pattern[:open]+alt+pattern[close+1:])...)
	}
	return expanded
}

// globToRegexp converts a glob pattern (supporting ** and brace alternation)
// into an anchored regular expression over slash-separated relative paths.
func globToRegexp(pattern string) (*regexp.Regexp, error) {
	pattern = filepath.ToSlash(pattern)
	variants := expandBraces(pattern)

	var alternatives []string
	for _, variant := range variants {
		var sb strings.Builder
		for i := 0; i < len(variant); i++ {
			c := variant[i]
			switch c {
			case '*':
				if i+1 < len(variant) && variant[i+1] == '*' {
					// "**/" or trailing "**" matches any number of path segments
					if i+2 < len(variant) && variant[i+2] == '/' {
						sb.WriteString(`(?:[^/]+/)*`)
						i += 2
					} else {
						sb.WriteString(`.*`)
						i++
					}
				} else {
					sb.WriteString(`[^/]*`)
				}
			case '?':
				sb.WriteString(`[^/]`)
			default:
				sb.WriteString(regexp.QuoteMeta(string(c)))
			}
		}
		alternatives = append(alternatives, sb.String())
	}

	return regexp.Compile(`^(?:` + strings.Join(alternatives, "|") + `)$`)
}

// isBinary heuristically detects binary content by scanning for NUL bytes.
func isBinary(data []byte) bool {
	probe := data
	if len(probe) > 8000 {
		probe = probe[:8000]
	}
	return bytes.IndexByte(probe, 0) != -1
}
