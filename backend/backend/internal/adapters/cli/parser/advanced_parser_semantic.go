package parser

import (
	"strings"
	"time"
)

// performSemanticAnalysis performs semantic analysis on the project
func (p *AdvancedTagParser) performSemanticAnalysis(analysis *ProjectAnalysis) {
	p.insights.LastAnalyzed = time.Now()

	// Detect code patterns
	p.detectCodePatterns(analysis)

	// Analyze module relationships
	p.analyzeModuleRelationships(analysis)

	// Detect architectural patterns
	p.detectArchitecturalPatterns(analysis)

	// Analyze tag usage patterns
	p.analyzeTagUsagePatterns(analysis)
}

// detectCodePatterns detects common code patterns in the project
func (p *AdvancedTagParser) detectCodePatterns(analysis *ProjectAnalysis) {
	// Repository pattern detection
	p.detectRepositoryPattern(analysis)

	// Service pattern detection
	p.detectServicePattern(analysis)

	// Handler pattern detection
	p.detectHandlerPattern(analysis)

	// Domain-driven design patterns
	p.detectDDDPatterns(analysis)

	// Dependency injection patterns
	p.detectDIPatterns(analysis)
}

// detectRepositoryPattern detects repository pattern usage
func (p *AdvancedTagParser) detectRepositoryPattern(analysis *ProjectAnalysis) {
	var repoFiles []string
	var repoCount int

	for _, tag := range analysis.Tags {
		if tag.Type == TagTypeRepository {
			repoCount++
			// Extract file path from context if available
			for _, module := range analysis.Modules {
				for _, moduleTag := range module.Tags {
					if moduleTag.Line == tag.Line {
						repoFiles = append(repoFiles, module.Files...)
						break
					}
				}
			}
		}
	}

	if repoCount > 0 {
		pattern := &CodePattern{
			Name:        "Repository Pattern",
			Type:        "architectural",
			Occurrences: repoCount,
			Files:       p.uniqueStrings(repoFiles),
			Confidence:  0.9,
			Metadata: map[string]string{
				"pattern_type": "data_access",
				"description":  "Repository pattern for data access abstraction",
			},
		}
		p.insights.Patterns["repository_pattern"] = pattern
	}
}

// detectServicePattern detects service layer pattern usage
func (p *AdvancedTagParser) detectServicePattern(analysis *ProjectAnalysis) {
	var serviceFiles []string
	var serviceCount int

	for _, tag := range analysis.Tags {
		if tag.Type == TagTypeService {
			serviceCount++
			for _, module := range analysis.Modules {
				for _, moduleTag := range module.Tags {
					if moduleTag.Line == tag.Line {
						serviceFiles = append(serviceFiles, module.Files...)
						break
					}
				}
			}
		}
	}

	if serviceCount > 0 {
		pattern := &CodePattern{
			Name:        "Service Layer Pattern",
			Type:        "architectural",
			Occurrences: serviceCount,
			Files:       p.uniqueStrings(serviceFiles),
			Confidence:  0.9,
			Metadata: map[string]string{
				"pattern_type": "business_logic",
				"description":  "Service layer for business logic encapsulation",
			},
		}
		p.insights.Patterns["service_pattern"] = pattern
	}
}

// detectHandlerPattern detects handler pattern usage
func (p *AdvancedTagParser) detectHandlerPattern(analysis *ProjectAnalysis) {
	var handlerFiles []string
	var handlerCount int

	for _, tag := range analysis.Tags {
		if tag.Type == TagTypeHandler {
			handlerCount++
		}
	}

	if handlerCount > 0 {
		pattern := &CodePattern{
			Name:        "Handler Pattern",
			Type:        "architectural",
			Occurrences: handlerCount,
			Files:       p.uniqueStrings(handlerFiles),
			Confidence:  0.85,
			Metadata: map[string]string{
				"pattern_type": "presentation",
				"description":  "Handler pattern for HTTP request processing",
			},
		}
		p.insights.Patterns["handler_pattern"] = pattern
	}
}

// detectDDDPatterns detects Domain-Driven Design patterns
func (p *AdvancedTagParser) detectDDDPatterns(analysis *ProjectAnalysis) {
	var domainCount int
	var aggregateCount int

	for _, tag := range analysis.Tags {
		if tag.Type == TagTypeDomain {
			domainCount++
		}

		// Check for aggregate hints in attributes
		if aggregate, exists := tag.Attributes["aggregate"]; exists && aggregate == "true" {
			aggregateCount++
		}
	}

	if domainCount > 0 {
		confidence := 0.7
		if aggregateCount > 0 {
			confidence = 0.9
		}

		pattern := &CodePattern{
			Name:        "Domain-Driven Design",
			Type:        "architectural",
			Occurrences: domainCount,
			Files:       []string{},
			Confidence:  confidence,
			Metadata: map[string]string{
				"pattern_type":    "domain_modeling",
				"domain_entities": string(rune(domainCount)),
				"aggregates":      string(rune(aggregateCount)),
				"description":     "Domain-driven design with domain entities and aggregates",
			},
		}
		p.insights.Patterns["ddd_pattern"] = pattern
	}
}

// detectDIPatterns detects Dependency Injection patterns
func (p *AdvancedTagParser) detectDIPatterns(analysis *ProjectAnalysis) {
	var providerCount int
	var moduleCount int

	for _, tag := range analysis.Tags {
		if tag.Type == TagTypeProvides {
			providerCount++
		}
		if tag.Type == TagTypeModule {
			moduleCount++
		}
	}

	if providerCount > 0 || moduleCount > 2 {
		pattern := &CodePattern{
			Name:        "Dependency Injection",
			Type:        "architectural",
			Occurrences: providerCount + moduleCount,
			Files:       []string{},
			Confidence:  0.8,
			Metadata: map[string]string{
				"pattern_type": "dependency_management",
				"providers":    string(rune(providerCount)),
				"modules":      string(rune(moduleCount)),
				"description":  "Dependency injection with providers and modules",
			},
		}
		p.insights.Patterns["di_pattern"] = pattern
	}
}

// analyzeModuleRelationships analyzes relationships between modules
func (p *AdvancedTagParser) analyzeModuleRelationships(analysis *ProjectAnalysis) {
	// Analyze coupling and cohesion
	for moduleName, module := range analysis.Modules {
		metrics := &ModuleMetrics{
			FileCount:       len(module.Files),
			TagCount:        len(module.Tags),
			DependencyCount: len(module.Dependencies),
		}

		// Calculate line count
		metrics.LineCount = p.calculateModuleLineCount(module)

		// Calculate coverage score (percentage of files with tags)
		taggedFiles := p.countTaggedFiles(module)
		if metrics.FileCount > 0 {
			metrics.CoverageScore = float64(taggedFiles) / float64(metrics.FileCount)
		}

		// Calculate quality score based on various factors
		metrics.QualityScore = p.calculateQualityScore(module, metrics)

		p.insights.Metrics.ModuleMetrics[moduleName] = metrics
	}
}

// detectArchitecturalPatterns detects overall architectural patterns
func (p *AdvancedTagParser) detectArchitecturalPatterns(analysis *ProjectAnalysis) {
	// Hexagonal Architecture detection
	if p.detectHexagonalArchitecture(analysis) {
		pattern := &CodePattern{
			Name:        "Hexagonal Architecture",
			Type:        "architectural",
			Occurrences: 1,
			Files:       []string{},
			Confidence:  0.8,
			Metadata: map[string]string{
				"pattern_type": "architectural_style",
				"description":  "Hexagonal architecture with clear separation of concerns",
			},
		}
		p.insights.Patterns["hexagonal_architecture"] = pattern
	}

	// Microservices pattern detection
	if p.detectMicroservicesPattern(analysis) {
		pattern := &CodePattern{
			Name:        "Microservices Pattern",
			Type:        "architectural",
			Occurrences: len(analysis.Modules),
			Files:       []string{},
			Confidence:  0.7,
			Metadata: map[string]string{
				"pattern_type": "architectural_style",
				"modules":      string(rune(len(analysis.Modules))),
				"description":  "Microservices architecture with modular design",
			},
		}
		p.insights.Patterns["microservices_pattern"] = pattern
	}
}

// detectHexagonalArchitecture detects hexagonal architecture patterns
func (p *AdvancedTagParser) detectHexagonalArchitecture(analysis *ProjectAnalysis) bool {
	hasHandlers := false
	hasServices := false
	hasRepositories := false
	hasDomain := false

	for _, tag := range analysis.Tags {
		switch tag.Type {
		case TagTypeHandler:
			hasHandlers = true
		case TagTypeService:
			hasServices = true
		case TagTypeRepository:
			hasRepositories = true
		case TagTypeDomain:
			hasDomain = true
		}
	}

	// Hexagonal architecture typically has all these layers
	return hasHandlers && hasServices && hasRepositories && hasDomain
}

// detectMicroservicesPattern detects microservices architecture
func (p *AdvancedTagParser) detectMicroservicesPattern(analysis *ProjectAnalysis) bool {
	// Consider it microservices if we have multiple independent modules
	// with their own handlers, services, and repositories
	moduleCount := 0

	for _, module := range analysis.Modules {
		hasHandler := false
		hasService := false
		hasRepository := false

		for _, tag := range module.Tags {
			switch tag.Type {
			case TagTypeHandler:
				hasHandler = true
			case TagTypeService:
				hasService = true
			case TagTypeRepository:
				hasRepository = true
			}
		}

		if hasHandler || hasService || hasRepository {
			moduleCount++
		}
	}

	return moduleCount >= 2
}

// analyzeTagUsagePatterns analyzes how tags are being used across the project
func (p *AdvancedTagParser) analyzeTagUsagePatterns(analysis *ProjectAnalysis) {
	tagDistribution := make(map[TagType]int)

	for _, tag := range analysis.Tags {
		tagDistribution[tag.Type]++
	}

	p.insights.Metrics.TagDistribution = tagDistribution

	// Analyze tag attribute usage
	p.analyzeTagAttributes(analysis)
}

// analyzeTagAttributes analyzes tag attribute usage patterns
func (p *AdvancedTagParser) analyzeTagAttributes(analysis *ProjectAnalysis) {
	attributeUsage := make(map[string]int)

	for _, tag := range analysis.Tags {
		for attrName := range tag.Attributes {
			attributeUsage[attrName]++
		}
	}

	// If certain attributes are used frequently, suggest patterns
	if attributeUsage["security"] > 2 {
		p.addRecommendation("security", "info",
			"Security attributes detected in multiple tags. Consider implementing a unified security framework.",
			"", 0,
			[]string{"Implement centralized security middleware", "Add security audit logging"})
	}

	if attributeUsage["observable"] > 2 {
		p.addRecommendation("observability", "info",
			"Observability attributes detected. Consider implementing comprehensive monitoring.",
			"", 0,
			[]string{"Add distributed tracing", "Implement metrics collection", "Set up health checks"})
	}
}

// calculateModuleLineCount calculates total lines of code for a module
func (p *AdvancedTagParser) calculateModuleLineCount(module *Module) int {
	// This would require reading files and counting lines
	// For now, return estimated count based on file count
	return len(module.Files) * 100 // Rough estimate
}

// countTaggedFiles counts how many files in a module have tags
func (p *AdvancedTagParser) countTaggedFiles(module *Module) int {
	taggedFiles := make(map[string]bool)

	for _, tag := range module.Tags {
		// This would require tracking which file each tag came from
		// For now, use a simple heuristic
		for _, file := range module.Files {
			if strings.Contains(file, strings.ToLower(string(tag.Type))) {
				taggedFiles[file] = true
				break
			}
		}
	}

	return len(taggedFiles)
}

// calculateQualityScore calculates a quality score for a module
func (p *AdvancedTagParser) calculateQualityScore(module *Module, metrics *ModuleMetrics) float64 {
	score := 0.0

	// Coverage contributes 40%
	score += metrics.CoverageScore * 0.4

	// Tag diversity contributes 30%
	tagTypes := make(map[TagType]bool)
	for _, tag := range module.Tags {
		tagTypes[tag.Type] = true
	}
	tagDiversity := float64(len(tagTypes)) / 5.0 // Normalize by max expected tag types
	if tagDiversity > 1.0 {
		tagDiversity = 1.0
	}
	score += tagDiversity * 0.3

	// Dependency management contributes 30%
	depScore := 1.0
	if metrics.DependencyCount > 5 {
		depScore = 0.5 // High dependency count reduces score
	} else if metrics.DependencyCount > 10 {
		depScore = 0.2
	}
	score += depScore * 0.3

	return score
}

// addRecommendation adds a recommendation to the insights
func (p *AdvancedTagParser) addRecommendation(recType, severity, message, file string, line int, suggestions []string) {
	recommendation := Recommendation{
		Type:        recType,
		Severity:    severity,
		Message:     message,
		File:        file,
		Line:        line,
		Suggestions: suggestions,
		Metadata:    make(map[string]string),
	}

	p.insights.Recommendations = append(p.insights.Recommendations, recommendation)
}
