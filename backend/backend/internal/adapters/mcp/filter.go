package mcp

import "strings"

// CommandFilter decides whether a Cobra command path should be exposed as an MCP tool.
type CommandFilter func(path []string) bool

// NewAllowDenyFilter constructs a filter from allow and deny lists.
// Patterns are defined using command segments separated by spaces (e.g. "migrate down").
// Deny rules always take precedence over allow rules.
func NewAllowDenyFilter(allowPatterns, denyPatterns []string) CommandFilter {
	allowed := normalizePatterns(allowPatterns)
	denied := normalizePatterns(denyPatterns)

	return func(path []string) bool {
		if len(path) == 0 {
			return false
		}
		for _, pattern := range denied {
			if matchPattern(pattern, path) {
				return false
			}
		}
		if len(allowed) == 0 {
			return true
		}
		for _, pattern := range allowed {
			if matchPattern(pattern, path) {
				return true
			}
		}
		return false
	}
}

func normalizePatterns(patterns []string) [][]string {
	result := make([][]string, 0, len(patterns))
	for _, raw := range patterns {
		if segments := parsePattern(raw); len(segments) > 0 {
			result = append(result, segments)
		}
	}
	return result
}

func parsePattern(value string) []string {
	cleaned := strings.TrimSpace(value)
	if cleaned == "" {
		return nil
	}
	replacer := strings.NewReplacer("/", " ", "_", " ")
	fields := strings.Fields(replacer.Replace(cleaned))
	normalized := make([]string, 0, len(fields))
	for _, field := range fields {
		normalized = append(normalized, strings.ToLower(strings.TrimSpace(field)))
	}
	return normalized
}

func matchPattern(pattern, path []string) bool {
	if len(pattern) == 0 || len(pattern) > len(path) {
		return false
	}
	for idx, segment := range pattern {
		if !strings.EqualFold(segment, path[idx]) {
			return false
		}
	}
	return true
}
