package parser

import (
	"context"
	"time"
)

// AdvancedTagParserIntegration integrates advanced parsing capabilities
type AdvancedTagParserIntegration struct {
	simpleParser   *TagParser
	advancedParser *AdvancedTagParser
	config         *AdvancedParserConfig
}

// NewAdvancedIntegration creates a new advanced integration
func NewAdvancedIntegration() *AdvancedTagParserIntegration {
	config := &AdvancedParserConfig{
		CacheEnabled:      true,
		CacheTTL:          time.Hour,
		CircularDetection: true,
		SemanticAnalysis:  true,
		MaxFileSize:       10 * 1024 * 1024, // 10MB
		IgnorePatterns:    []string{"vendor/", ".git/", "_test.go", "testdata/"},
	}

	return &AdvancedTagParserIntegration{
		simpleParser:   NewTagParser(NewNullCache()),
		advancedParser: NewAdvancedTagParser(config),
		config:         config,
	}
}

// AnalyzeProjectWithInsights performs comprehensive project analysis with insights
func (i *AdvancedTagParserIntegration) AnalyzeProjectWithInsights(projectPath string) (*ProjectAnalysis, *SemanticInsights, *DependencyGraph, error) {
	ctx := context.Background()

	// Perform advanced analysis
	analysis, err := i.advancedParser.AnalyzeProjectAdvanced(ctx, projectPath)
	if err != nil {
		return nil, nil, nil, err
	}

	return analysis, i.advancedParser.insights, i.advancedParser.depGraph, nil
}

// AnalyzeProjectSimple performs simple analysis (fallback)
func (i *AdvancedTagParserIntegration) AnalyzeProjectSimple(projectPath string) (*ProjectAnalysis, error) {
	return i.simpleParser.AnalyzeProject(projectPath)
}

// GetRecommendations returns actionable recommendations
func (i *AdvancedTagParserIntegration) GetRecommendations() []Recommendation {
	if i.advancedParser.insights != nil {
		return i.advancedParser.insights.Recommendations
	}
	return []Recommendation{}
}

// GetCodePatterns returns detected code patterns
func (i *AdvancedTagParserIntegration) GetCodePatterns() map[string]*CodePattern {
	if i.advancedParser.insights != nil {
		return i.advancedParser.insights.Patterns
	}
	return make(map[string]*CodePattern)
}

// GetProjectMetrics returns project-level metrics
func (i *AdvancedTagParserIntegration) GetProjectMetrics() *ProjectMetrics {
	if i.advancedParser.insights != nil {
		return i.advancedParser.insights.Metrics
	}
	return &ProjectMetrics{}
}

// GetCircularDependencies returns detected circular dependencies
func (i *AdvancedTagParserIntegration) GetCircularDependencies() [][]string {
	if i.advancedParser.depGraph != nil {
		return i.advancedParser.depGraph.Cycles
	}
	return [][]string{}
}

// ConfigureAdvancedFeatures configures advanced parsing features
func (i *AdvancedTagParserIntegration) ConfigureAdvancedFeatures(config *AdvancedParserConfig) {
	i.config = config
	i.advancedParser = NewAdvancedTagParser(config)
}

// ClearCache clears the parser cache
func (i *AdvancedTagParserIntegration) ClearCache() error {
	return i.advancedParser.cache.Clear()
}
