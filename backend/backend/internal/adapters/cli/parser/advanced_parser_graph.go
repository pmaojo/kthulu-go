package parser

import (
	"fmt"
)

// buildDependencyGraph builds the project dependency graph
func (p *AdvancedTagParser) buildDependencyGraph(analysis *ProjectAnalysis) {
	// Create nodes for each module
	for moduleName, module := range analysis.Modules {
		node := &DependencyNode{
			ID:       moduleName,
			Name:     moduleName,
			Type:     "module",
			Tags:     module.Tags,
			Metadata: make(map[string]string),
		}

		// Add metadata from tags
		for _, tag := range module.Tags {
			if tag.Type == TagTypeModule {
				for key, value := range tag.Attributes {
					node.Metadata[key] = value
				}
			}
		}

		p.depGraph.Nodes[moduleName] = node
	}

	// Create edges for dependencies
	for _, dep := range analysis.Dependencies {
		edge := &DependencyEdge{
			From:     dep.From,
			To:       dep.To,
			Type:     dep.Type,
			Weight:   1,
			Metadata: make(map[string]string),
		}

		p.depGraph.Edges = append(p.depGraph.Edges, edge)

		// Update node degrees
		if fromNode, exists := p.depGraph.Nodes[dep.From]; exists {
			fromNode.OutDegree++
		}
		if toNode, exists := p.depGraph.Nodes[dep.To]; exists {
			toNode.InDegree++
		}
	}

	// Calculate dependency levels (topological ordering)
	p.calculateDependencyLevels()

	// Add symbol-level dependencies
	p.addSymbolDependencies(analysis)
}

// calculateDependencyLevels calculates the dependency levels for topological ordering
func (p *AdvancedTagParser) calculateDependencyLevels() {
	visited := make(map[string]bool)
	levels := make(map[string]int)

	// Calculate levels using DFS
	var calculateLevel func(nodeID string) int
	calculateLevel = func(nodeID string) int {
		if visited[nodeID] {
			return levels[nodeID]
		}

		visited[nodeID] = true
		maxLevel := 0

		// Check all dependencies
		for _, edge := range p.depGraph.Edges {
			if edge.From == nodeID {
				depLevel := calculateLevel(edge.To)
				if depLevel >= maxLevel {
					maxLevel = depLevel + 1
				}
			}
		}

		levels[nodeID] = maxLevel
		return maxLevel
	}

	// Calculate levels for all nodes
	for nodeID := range p.depGraph.Nodes {
		calculateLevel(nodeID)
	}

	p.depGraph.Levels = levels
}

// addSymbolDependencies adds symbol-level dependencies to the graph
func (p *AdvancedTagParser) addSymbolDependencies(analysis *ProjectAnalysis) {
	// Add symbol nodes
	for moduleName := range analysis.Modules {
		// This would require analyzing file contents for symbol usage
		// For now, we'll add placeholder symbol dependencies
		symbolID := fmt.Sprintf("%s:symbol", moduleName)

		if _, exists := p.depGraph.Nodes[symbolID]; !exists {
			symbolNode := &DependencyNode{
				ID:        symbolID,
				Name:      fmt.Sprintf("%s symbols", moduleName),
				Type:      "symbol_group",
				Tags:      []Tag{},
				Metadata:  map[string]string{"module": moduleName},
				InDegree:  0,
				OutDegree: 0,
			}
			p.depGraph.Nodes[symbolID] = symbolNode
		}
	}
}

// detectCircularDependencies detects circular dependencies in the graph
func (p *AdvancedTagParser) detectCircularDependencies() {
	visited := make(map[string]int) // 0: white, 1: gray, 2: black
	var path []string
	var cycles [][]string

	var dfs func(nodeID string) bool
	dfs = func(nodeID string) bool {
		if visited[nodeID] == 1 {
			// Gray node - we found a back edge, indicating a cycle
			cycleStart := -1
			for i, node := range path {
				if node == nodeID {
					cycleStart = i
					break
				}
			}

			if cycleStart != -1 {
				cycle := make([]string, len(path)-cycleStart)
				copy(cycle, path[cycleStart:])
				cycle = append(cycle, nodeID) // Complete the cycle
				cycles = append(cycles, cycle)
			}
			return true
		}

		if visited[nodeID] == 2 {
			// Black node - already processed
			return false
		}

		// Mark as gray (being processed)
		visited[nodeID] = 1
		path = append(path, nodeID)

		// Visit all adjacent nodes
		hasCycle := false
		for _, edge := range p.depGraph.Edges {
			if edge.From == nodeID {
				if dfs(edge.To) {
					hasCycle = true
				}
			}
		}

		// Mark as black (completed)
		visited[nodeID] = 2
		path = path[:len(path)-1] // Remove from path

		return hasCycle
	}

	// Check all nodes for cycles
	for nodeID := range p.depGraph.Nodes {
		if visited[nodeID] == 0 {
			dfs(nodeID)
		}
	}

	// Remove duplicate cycles
	p.depGraph.Cycles = p.uniqueCycles(cycles)

	// Generate recommendations for cycles
	p.generateCycleRecommendations()
}

// uniqueCycles removes duplicate cycles
func (p *AdvancedTagParser) uniqueCycles(cycles [][]string) [][]string {
	seen := make(map[string]bool)
	var unique [][]string

	for _, cycle := range cycles {
		// Normalize cycle representation
		normalized := p.normalizeCycle(cycle)
		key := fmt.Sprintf("%v", normalized)

		if !seen[key] {
			seen[key] = true
			unique = append(unique, cycle)
		}
	}

	return unique
}

// normalizeCycle normalizes a cycle to a canonical form
func (p *AdvancedTagParser) normalizeCycle(cycle []string) []string {
	if len(cycle) == 0 {
		return cycle
	}

	// Find the lexicographically smallest starting point
	minIndex := 0
	for i, node := range cycle {
		if node < cycle[minIndex] {
			minIndex = i
		}
	}

	// Rotate the cycle to start from the minimum element
	normalized := make([]string, len(cycle))
	for i := 0; i < len(cycle); i++ {
		normalized[i] = cycle[(minIndex+i)%len(cycle)]
	}

	return normalized
}

// generateCycleRecommendations generates recommendations for breaking cycles
func (p *AdvancedTagParser) generateCycleRecommendations() {
	for _, cycle := range p.depGraph.Cycles {
		message := fmt.Sprintf("Circular dependency detected: %s",
			fmt.Sprintf("%v", cycle))

		suggestions := []string{
			"Extract common functionality into a shared module",
			"Use dependency inversion principle",
			"Consider using interfaces to break direct dependencies",
			"Refactor to use event-driven communication",
		}

		// Add specific suggestions based on cycle characteristics
		if len(cycle) == 2 {
			suggestions = append(suggestions, "Consider merging these two modules if they're tightly coupled")
		} else if len(cycle) > 4 {
			suggestions = append(suggestions, "This cycle is complex - consider major architectural refactoring")
		}

		p.addRecommendation("circular_dependency", "warning", message, "", 0, suggestions)
	}
}

// generateRecommendations generates additional recommendations based on analysis
func (p *AdvancedTagParser) generateRecommendations() {
	// Module recommendations
	p.generateModuleRecommendations()

	// Architecture recommendations
	p.generateArchitectureRecommendations()

	// Tag usage recommendations
	p.generateTagUsageRecommendations()

	// Performance recommendations
	p.generatePerformanceRecommendations()
}

// generateModuleRecommendations generates module-specific recommendations
func (p *AdvancedTagParser) generateModuleRecommendations() {
	for moduleName, metrics := range p.insights.Metrics.ModuleMetrics {
		// Low coverage recommendation
		if metrics.CoverageScore < 0.5 {
			p.addRecommendation("module_coverage", "info",
				fmt.Sprintf("Module '%s' has low tag coverage (%.1f%%). Consider adding more @kthulu tags for better code organization.",
					moduleName, metrics.CoverageScore*100),
				"", 0,
				[]string{
					"Add @kthulu:service tags to service files",
					"Add @kthulu:repository tags to repository files",
					"Add @kthulu:handler tags to handler files",
				})
		}

		// High dependency count
		if metrics.DependencyCount > 8 {
			p.addRecommendation("module_coupling", "warning",
				fmt.Sprintf("Module '%s' has high coupling with %d dependencies. Consider reducing dependencies.",
					moduleName, metrics.DependencyCount),
				"", 0,
				[]string{
					"Extract shared functionality into common modules",
					"Use dependency inversion to reduce direct coupling",
					"Consider breaking down the module into smaller parts",
				})
		}

		// Quality score recommendations
		if metrics.QualityScore < 0.6 {
			p.addRecommendation("module_quality", "warning",
				fmt.Sprintf("Module '%s' has low quality score (%.1f%%). Consider improvements.",
					moduleName, metrics.QualityScore*100),
				"", 0,
				[]string{
					"Improve tag coverage",
					"Reduce dependencies",
					"Add comprehensive documentation",
				})
		}
	}
}

// generateArchitectureRecommendations generates architecture-level recommendations
func (p *AdvancedTagParser) generateArchitectureRecommendations() {
	// Check for missing layers
	hasHandlers := false
	hasServices := false
	hasRepositories := false
	hasDomain := false

	for tagType, count := range p.insights.Metrics.TagDistribution {
		switch tagType {
		case TagTypeHandler:
			hasHandlers = count > 0
		case TagTypeService:
			hasServices = count > 0
		case TagTypeRepository:
			hasRepositories = count > 0
		case TagTypeDomain:
			hasDomain = count > 0
		}
	}

	var missingLayers []string
	if !hasHandlers {
		missingLayers = append(missingLayers, "handlers")
	}
	if !hasServices {
		missingLayers = append(missingLayers, "services")
	}
	if !hasRepositories {
		missingLayers = append(missingLayers, "repositories")
	}
	if !hasDomain {
		missingLayers = append(missingLayers, "domain models")
	}

	if len(missingLayers) > 0 {
		p.addRecommendation("architecture_completeness", "info",
			fmt.Sprintf("Consider implementing missing architectural layers: %s",
				fmt.Sprintf("%v", missingLayers)),
			"", 0,
			[]string{
				"Implement complete layered architecture",
				"Add missing @kthulu tags to identify layers",
				"Follow hexagonal architecture principles",
			})
	}
}

// generateTagUsageRecommendations generates tag usage recommendations
func (p *AdvancedTagParser) generateTagUsageRecommendations() {
	totalTags := 0
	for _, count := range p.insights.Metrics.TagDistribution {
		totalTags += count
	}

	if totalTags == 0 {
		p.addRecommendation("tag_usage", "warning",
			"No @kthulu tags found. Consider adding tags for better code organization and automation.",
			"", 0,
			[]string{
				"Start with @kthulu:module tags",
				"Add @kthulu:service, @kthulu:repository, @kthulu:handler tags",
				"Use attributes for advanced features (security=high, observable=true)",
			})
	} else if totalTags < 5 {
		p.addRecommendation("tag_usage", "info",
			fmt.Sprintf("Only %d @kthulu tags found. Consider adding more tags for comprehensive automation.", totalTags),
			"", 0,
			[]string{
				"Add tags to all architectural layers",
				"Use attributes for configuration",
				"Consider @kthulu:observable and @kthulu:security tags",
			})
	}
}

// generatePerformanceRecommendations generates performance-related recommendations
func (p *AdvancedTagParser) generatePerformanceRecommendations() {
	// Check dependency graph depth
	maxDepth := 0
	for _, level := range p.depGraph.Levels {
		if level > maxDepth {
			maxDepth = level
		}
	}

	if maxDepth > 5 {
		p.addRecommendation("dependency_depth", "warning",
			fmt.Sprintf("Dependency chain is deep (%d levels). This may impact build times and understanding.", maxDepth),
			"", 0,
			[]string{
				"Flatten dependency hierarchy",
				"Extract common dependencies",
				"Consider dependency injection to reduce coupling",
			})
	}

	// Check for highly connected modules
	for nodeID, node := range p.depGraph.Nodes {
		totalConnections := node.InDegree + node.OutDegree
		if totalConnections > 10 {
			p.addRecommendation("module_connectivity", "warning",
				fmt.Sprintf("Module '%s' is highly connected (%d connections). Consider refactoring.",
					nodeID, totalConnections),
				"", 0,
				[]string{
					"Split module into smaller, focused modules",
					"Use facade pattern to simplify interfaces",
					"Extract common functionality",
				})
		}
	}
}

// calculateMetrics calculates final project metrics
func (p *AdvancedTagParser) calculateMetrics(analysis *ProjectAnalysis) {
	metrics := p.insights.Metrics

	// Basic counts
	metrics.ModuleCount = len(analysis.Modules)
	metrics.CycleCount = len(p.depGraph.Cycles)

	// Calculate total files and tagged files
	allFiles := make(map[string]bool)
	taggedFiles := make(map[string]bool)

	for _, module := range analysis.Modules {
		for _, file := range module.Files {
			allFiles[file] = true
		}

		if len(module.Tags) > 0 {
			for _, file := range module.Files {
				taggedFiles[file] = true
			}
		}
	}

	metrics.TotalFiles = len(allFiles)
	metrics.TaggedFiles = len(taggedFiles)

	// Calculate complexity score
	metrics.ComplexityScore = p.calculateComplexityScore(analysis)

	// Calculate dependency depth
	maxDepth := 0
	for _, level := range p.depGraph.Levels {
		if level > maxDepth {
			maxDepth = level
		}
	}
	metrics.DependencyDepth = maxDepth
}

// calculateComplexityScore calculates a complexity score for the project
func (p *AdvancedTagParser) calculateComplexityScore(analysis *ProjectAnalysis) float64 {
	score := 0.0

	// Module count contributes to complexity
	moduleComplexity := float64(len(analysis.Modules)) * 0.1
	if moduleComplexity > 1.0 {
		moduleComplexity = 1.0
	}
	score += moduleComplexity * 0.3

	// Dependency count
	depComplexity := float64(len(analysis.Dependencies)) * 0.05
	if depComplexity > 1.0 {
		depComplexity = 1.0
	}
	score += depComplexity * 0.3

	// Cycle count (higher is worse)
	cycleComplexity := float64(len(p.depGraph.Cycles)) * 0.2
	if cycleComplexity > 1.0 {
		cycleComplexity = 1.0
	}
	score += cycleComplexity * 0.2

	// Dependency depth
	depthComplexity := float64(p.insights.Metrics.DependencyDepth) * 0.1
	if depthComplexity > 1.0 {
		depthComplexity = 1.0
	}
	score += depthComplexity * 0.2

	return score
}
