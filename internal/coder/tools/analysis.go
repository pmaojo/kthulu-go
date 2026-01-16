package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
)

// AnalysisTool provides deep project analysis and insights
var AnalysisTool = &Tool{
	Name: "analysis",
	Description: `Analyze the project structure, code quality, architecture patterns, and metrics.
Returns a summary of modules, complexity, patterns (like DDD/Clean Architecture usage), and improvement recommendations.

Use this when:
- The user asks to "review" or "optimize" the code.
- You need to understand the project's overall health or structure.
- You are looking for security or performance issues.`,
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to analyze (defaults to current directory '.')",
			},
		},
	},
	Execute: executeAnalysis,
}

func executeAnalysis(ctx context.Context, args map[string]any) (*Result, error) {
	path := "."
	if p, ok := args["path"].(string); ok && p != "" {
		path = p
	}

	// Use advanced integration for context analysis
	integration := parser.NewAdvancedIntegration()
	result, insights, _, err := integration.AnalyzeProjectWithInsights(path)
	if err != nil {
		return &Result{
			Success: false,
			Error:   fmt.Sprintf("Analysis failed: %v", err),
		}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Project Analysis Report (%s)\n\n", path))

	// 1. Modules
	if len(result.Modules) > 0 {
		sb.WriteString("## Modules\n")
		for _, mod := range result.Modules {
			sb.WriteString(fmt.Sprintf("- **%s**: %d files\n", mod.Name, len(mod.Files)))
		}
		sb.WriteString("\n")
	}

	// 2. Metrics
	metrics := integration.GetProjectMetrics()
	if metrics != nil {
		sb.WriteString("## Metrics\n")
		sb.WriteString(fmt.Sprintf("- Total Files: %d\n", metrics.TotalFiles))
		sb.WriteString(fmt.Sprintf("- Complexity Score: %.2f\n", metrics.ComplexityScore))
		sb.WriteString("\n")
	}

	// 3. Architecture Patterns
	if insights != nil && len(insights.Patterns) > 0 {
		sb.WriteString("## Detected Patterns\n")
		for _, p := range insights.Patterns {
			if p.Confidence > 0.5 {
				sb.WriteString(fmt.Sprintf("- %s (%.0f%% confidence)\n", p.Name, p.Confidence*100))
			}
		}
		sb.WriteString("\n")
	}

	// 4. Recommendations
	recommendations := integration.GetRecommendations()
	if len(recommendations) > 0 {
		sb.WriteString("## AI Recommendations\n")
		for _, rec := range recommendations {
			icon := "ℹ️"
			if strings.Contains(strings.ToLower(rec.Type), "security") {
				icon = "🔒"
			} else if strings.Contains(strings.ToLower(rec.Type), "performance") {
				icon = "⚡"
			}
			sb.WriteString(fmt.Sprintf("- %s **%s** (%s): %s\n", icon, rec.Type, rec.Severity, rec.Message))
		}
		sb.WriteString("\n")
	}

	output := sb.String()

	// Create a shorter display summary for TUI
	patternsCount := 0
	if insights != nil {
		patternsCount = len(insights.Patterns)
	}

	display := fmt.Sprintf("Analyzed %d modules. Found %d patterns and %d recommendations.",
		len(result.Modules),
		patternsCount,
		len(recommendations))

	return &Result{
		Success: true,
		Output:  output,
		Display: display,
	}, nil
}
