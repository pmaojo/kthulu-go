package resolver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/parser"
)

// DependencyResolver resolves module dependencies intelligently
type DependencyResolver struct {
	modules      map[string]*parser.Module
	dependencies []parser.Dependency
	rules        map[string][]string // module -> required dependencies
}

// ResolutionPlan represents a dependency resolution plan
type ResolutionPlan struct {
	RequiredModules []string         `json:"required_modules"`
	InstallOrder    []string         `json:"install_order"`
	OptionalModules []string         `json:"optional_modules"`
	Conflicts       []ConflictInfo   `json:"conflicts,omitempty"`
	Warnings        []string         `json:"warnings,omitempty"`
	Recommendations []Recommendation `json:"recommendations,omitempty"`
}

// ConflictInfo represents a dependency conflict
type ConflictInfo struct {
	Type        string   `json:"type"` // circular, incompatible, version
	Modules     []string `json:"modules"`
	Description string   `json:"description"`
	Suggestions []string `json:"suggestions"`
}

// Recommendation represents an intelligent recommendation
type Recommendation struct {
	Type      string `json:"type"` // add, remove, upgrade, configure
	Module    string `json:"module"`
	Reason    string `json:"reason"`
	Impact    string `json:"impact"` // low, medium, high
	AutoApply bool   `json:"auto_apply"`
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver(analysis *parser.ProjectAnalysis) *DependencyResolver {
	resolver := &DependencyResolver{
		modules:      analysis.Modules,
		dependencies: analysis.Dependencies,
		rules:        make(map[string][]string),
	}

	// Initialize dependency rules
	resolver.initializeRules()

	return resolver
}

// initializeRules initializes the dependency rules
func (r *DependencyResolver) initializeRules() {
	// Core dependencies - modules that are required by other modules
	r.rules["auth"] = []string{"user"}
	r.rules["organization"] = []string{"user", "auth"}
	r.rules["contact"] = []string{"user", "organization"}
	r.rules["product"] = []string{"user", "organization"}
	r.rules["invoice"] = []string{"user", "organization", "product", "contact"}
	r.rules["payment"] = []string{"user", "invoice"}
	r.rules["inventory"] = []string{"user", "organization", "product"}
	r.rules["calendar"] = []string{"user", "organization", "contact"}
	r.rules["verifactu"] = []string{"invoice", "organization"}
	r.rules["oauthsso"] = []string{"user", "auth"}

	// Advanced modules
	r.rules["realtime"] = []string{"user", "auth"}
	r.rules["audit"] = []string{"user"}
	r.rules["notification"] = []string{"user"}
}

// ResolveDependencies resolves dependencies for the given modules
func (r *DependencyResolver) ResolveDependencies(requestedModules []string) (*ResolutionPlan, error) {
	fmt.Printf("🧠 Resolving dependencies for modules: %s\n", strings.Join(requestedModules, ", "))

	plan := &ResolutionPlan{
		RequiredModules: []string{},
		InstallOrder:    []string{},
		OptionalModules: []string{},
		Conflicts:       []ConflictInfo{},
		Warnings:        []string{},
		Recommendations: []Recommendation{},
	}

	// Step 1: Resolve direct dependencies
	allRequired := make(map[string]bool)
	for _, module := range requestedModules {
		r.resolveDependenciesRecursive(module, allRequired, plan)
	}

	// Convert map to slice
	for module := range allRequired {
		plan.RequiredModules = append(plan.RequiredModules, module)
	}

	// Step 2: Calculate installation order (topological sort)
	installOrder, err := r.calculateInstallOrder(plan.RequiredModules)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate install order: %w", err)
	}
	plan.InstallOrder = installOrder

	// Step 3: Detect conflicts
	r.detectConflicts(plan)

	// Step 4: Generate recommendations
	r.generateRecommendations(requestedModules, plan)

	// Step 5: Suggest optional modules
	r.suggestOptionalModules(plan)

	fmt.Printf("✅ Resolution complete: %d required, %d optional, %d conflicts\n",
		len(plan.RequiredModules), len(plan.OptionalModules), len(plan.Conflicts))

	return plan, nil
}

// resolveDependenciesRecursive recursively resolves dependencies for a module
func (r *DependencyResolver) resolveDependenciesRecursive(module string, resolved map[string]bool, plan *ResolutionPlan) {
	if resolved[module] {
		return // Already resolved
	}

	resolved[module] = true

	// Get dependencies for this module
	deps, exists := r.rules[module]
	if !exists {
		// No dependencies defined - this is a leaf module
		return
	}

	// Resolve dependencies recursively
	for _, dep := range deps {
		r.resolveDependenciesRecursive(dep, resolved, plan)
	}
}

// calculateInstallOrder calculates the correct installation order using topological sort
func (r *DependencyResolver) calculateInstallOrder(modules []string) ([]string, error) {
	// Build adjacency list
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize
	for _, module := range modules {
		graph[module] = []string{}
		inDegree[module] = 0
	}

	// Build graph based on dependency rules
	for _, module := range modules {
		if deps, exists := r.rules[module]; exists {
			for _, dep := range deps {
				if _, depExists := inDegree[dep]; depExists {
					graph[dep] = append(graph[dep], module)
					inDegree[module]++
				}
			}
		}
	}

	// Topological sort using Kahn's algorithm
	var result []string
	queue := []string{}

	// Find nodes with no incoming edges
	for module, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, module)
		}
	}

	for len(queue) > 0 {
		// Remove node from queue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Remove edges from current node
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles
	if len(result) != len(modules) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

// detectConflicts detects various types of conflicts
func (r *DependencyResolver) detectConflicts(plan *ResolutionPlan) {
	// Check for circular dependencies (already handled in topological sort)

	// Check for incompatible modules
	r.checkIncompatibleModules(plan)

	// Check for missing core dependencies
	r.checkMissingCoreDependencies(plan)
}

// checkIncompatibleModules checks for known incompatible module combinations
func (r *DependencyResolver) checkIncompatibleModules(plan *ResolutionPlan) {
	incompatiblePairs := map[string][]string{
		"sqlite":     {"mysql", "postgresql"},
		"mysql":      {"postgresql"},
		"local_auth": {"oauthsso"},
		// Add more incompatible combinations
	}

	moduleSet := make(map[string]bool)
	for _, module := range plan.RequiredModules {
		moduleSet[module] = true
	}

	for module, incompatibles := range incompatiblePairs {
		if moduleSet[module] {
			for _, incompatible := range incompatibles {
				if moduleSet[incompatible] {
					conflict := ConflictInfo{
						Type:        "incompatible",
						Modules:     []string{module, incompatible},
						Description: fmt.Sprintf("Modules '%s' and '%s' are incompatible", module, incompatible),
						Suggestions: []string{
							fmt.Sprintf("Choose either '%s' or '%s', not both", module, incompatible),
							"Consider using a different approach that supports both",
						},
					}
					plan.Conflicts = append(plan.Conflicts, conflict)
				}
			}
		}
	}
}

// checkMissingCoreDependencies checks for missing core dependencies
func (r *DependencyResolver) checkMissingCoreDependencies(plan *ResolutionPlan) {
	// Ensure core modules are present when needed
	coreModules := []string{"user", "auth"}
	hasCore := false

	for _, module := range plan.RequiredModules {
		for _, core := range coreModules {
			if module == core {
				hasCore = true
				break
			}
		}
		if hasCore {
			break
		}
	}

	if !hasCore && len(plan.RequiredModules) > 0 {
		plan.Warnings = append(plan.Warnings,
			"No core authentication modules detected - consider adding 'user' and 'auth' modules")
	}
}

// generateRecommendations generates intelligent recommendations
func (r *DependencyResolver) generateRecommendations(requestedModules []string, plan *ResolutionPlan) {
	// Recommend security enhancements
	if contains(requestedModules, "payment") || contains(requestedModules, "invoice") {
		if !contains(plan.RequiredModules, "audit") {
			rec := Recommendation{
				Type:      "add",
				Module:    "audit",
				Reason:    "Financial modules benefit from audit logging for compliance",
				Impact:    "medium",
				AutoApply: false,
			}
			plan.Recommendations = append(plan.Recommendations, rec)
		}
	}

	// Recommend performance monitoring
	if len(plan.RequiredModules) > 3 {
		rec := Recommendation{
			Type:      "configure",
			Module:    "observability",
			Reason:    "Multiple modules benefit from centralized monitoring",
			Impact:    "high",
			AutoApply: true,
		}
		plan.Recommendations = append(plan.Recommendations, rec)
	}

	// Recommend real-time features for interactive modules
	if contains(requestedModules, "chat") || contains(requestedModules, "collaboration") {
		if !contains(plan.RequiredModules, "realtime") {
			rec := Recommendation{
				Type:      "add",
				Module:    "realtime",
				Reason:    "Interactive modules work better with real-time capabilities",
				Impact:    "high",
				AutoApply: false,
			}
			plan.Recommendations = append(plan.Recommendations, rec)
		}
	}
}

// suggestOptionalModules suggests optional modules that might be useful
func (r *DependencyResolver) suggestOptionalModules(plan *ResolutionPlan) {
	suggestions := map[string][]string{
		"user":         {"notification", "audit"},
		"organization": {"contact", "calendar"},
		"product":      {"inventory", "pricing"},
		"invoice":      {"payment", "verifactu"},
		"contact":      {"calendar", "communication"},
	}

	for _, module := range plan.RequiredModules {
		if optionals, exists := suggestions[module]; exists {
			for _, optional := range optionals {
				if !contains(plan.RequiredModules, optional) && !contains(plan.OptionalModules, optional) {
					plan.OptionalModules = append(plan.OptionalModules, optional)
				}
			}
		}
	}

	// Sort for consistent output
	sort.Strings(plan.OptionalModules)
}

// GetModuleInfo returns detailed information about a module
func (r *DependencyResolver) GetModuleInfo(moduleName string) (*ModuleInfo, error) {
	module, exists := r.modules[moduleName]
	if !exists {
		return nil, fmt.Errorf("module '%s' not found", moduleName)
	}

	info := &ModuleInfo{
		Name:         module.Name,
		Package:      module.Package,
		Dependencies: module.Dependencies,
		Description:  r.getModuleDescription(moduleName),
		Category:     r.getModuleCategory(moduleName),
		Complexity:   r.calculateComplexity(module),
		EstimatedLOC: r.estimateLinesOfCode(module),
		Tags:         len(module.Tags),
	}

	return info, nil
}

// ModuleInfo represents detailed module information
type ModuleInfo struct {
	Name         string   `json:"name"`
	Package      string   `json:"package"`
	Dependencies []string `json:"dependencies"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Complexity   string   `json:"complexity"`
	EstimatedLOC int      `json:"estimated_loc"`
	Tags         int      `json:"tags"`
}

// Helper functions
func (r *DependencyResolver) getModuleDescription(module string) string {
	descriptions := map[string]string{
		"user":         "User management and authentication core",
		"auth":         "Authentication and authorization system",
		"organization": "Multi-tenant organization management",
		"contact":      "Customer and vendor contact management",
		"product":      "Product catalog and management",
		"invoice":      "Invoice generation and management",
		"payment":      "Payment processing and gateway integration",
		"inventory":    "Inventory and warehouse management",
		"calendar":     "Scheduling and calendar management",
		"verifactu":    "Spanish fiscal compliance (VeriFACTU)",
		"oauthsso":     "OAuth and SSO integration",
		"realtime":     "Real-time communication and WebSocket support",
		"audit":        "Audit logging and compliance tracking",
		"notification": "Multi-channel notification system",
	}

	if desc, exists := descriptions[module]; exists {
		return desc
	}
	return "Custom module"
}

func (r *DependencyResolver) getModuleCategory(module string) string {
	categories := map[string]string{
		"user":         "Core",
		"auth":         "Core",
		"organization": "Core",
		"contact":      "Business",
		"product":      "Business",
		"invoice":      "Business",
		"payment":      "Integration",
		"inventory":    "Business",
		"calendar":     "Business",
		"verifactu":    "Compliance",
		"oauthsso":     "Integration",
		"realtime":     "Infrastructure",
		"audit":        "Compliance",
		"notification": "Infrastructure",
	}

	if cat, exists := categories[module]; exists {
		return cat
	}
	return "Custom"
}

func (r *DependencyResolver) calculateComplexity(module *parser.Module) string {
	// Simple complexity calculation based on dependencies and files
	score := len(module.Dependencies)*2 + len(module.Files)

	if score < 5 {
		return "Low"
	} else if score < 15 {
		return "Medium"
	}
	return "High"
}

func (r *DependencyResolver) estimateLinesOfCode(module *parser.Module) int {
	// Rough estimation based on number of files and complexity
	baseLines := len(module.Files) * 150      // Average lines per file
	depBonus := len(module.Dependencies) * 50 // Additional complexity
	return baseLines + depBonus
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
