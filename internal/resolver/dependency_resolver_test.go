package resolver

import (
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/stretchr/testify/assert"
)

func TestGetModuleInfo(t *testing.T) {
	// Setup
	analysis := &parser.ProjectAnalysis{
		Modules: map[string]*parser.Module{
			"user": {
				Name:    "user",
				Package: "user",
				Files:   []string{"user.go"},
			},
		},
	}
	r := NewDependencyResolver(analysis)

	// Test correct info retrieval
	info, err := r.GetModuleInfo("user")
	assert.NoError(t, err)
	assert.Equal(t, "user", info.Name)
	assert.Equal(t, "User management and authentication core", info.Description)
	assert.Equal(t, "Core", info.Category)

	// Test missing module
	_, err = r.GetModuleInfo("missing")
	assert.Error(t, err)

	// Test custom module description/category
	analysis.Modules["custom"] = &parser.Module{
		Name:    "custom",
		Package: "custom",
		Files:   []string{"custom.go"},
	}
	// Re-create resolver because modules are passed in constructor
	r = NewDependencyResolver(analysis)

	info, err = r.GetModuleInfo("custom")
	assert.NoError(t, err)
	assert.Equal(t, "Custom module", info.Description)
	assert.Equal(t, "Custom", info.Category)
}

func TestResolveDependencies(t *testing.T) {
	analysis := &parser.ProjectAnalysis{
		Modules:      map[string]*parser.Module{},
		Dependencies: []parser.Dependency{},
	}
	r := NewDependencyResolver(analysis)

	// Test basic resolution
	plan, err := r.ResolveDependencies([]string{"auth"})
	assert.NoError(t, err)
	assert.Contains(t, plan.RequiredModules, "auth")
	assert.Contains(t, plan.RequiredModules, "user") // auth depends on user
}

func TestCheckIncompatibleModules(t *testing.T) {
	analysis := &parser.ProjectAnalysis{
		Modules:      map[string]*parser.Module{},
		Dependencies: []parser.Dependency{},
	}
	r := NewDependencyResolver(analysis)

	// Test incompatibility
	// We need to force incompatible modules into the plan
	// ResolveDependencies calculates install order, then checks conflicts

	// sqlite and mysql are incompatible
	plan, err := r.ResolveDependencies([]string{"sqlite", "mysql"})
	assert.NoError(t, err) // It doesn't return error, but adds conflicts to plan

	hasConflict := false
	for _, conflict := range plan.Conflicts {
		if conflict.Type == "incompatible" {
			hasConflict = true
			break
		}
	}
	assert.True(t, hasConflict, "Should detect incompatible modules")
}

func TestSuggestOptionalModules(t *testing.T) {
	analysis := &parser.ProjectAnalysis{
		Modules:      map[string]*parser.Module{},
		Dependencies: []parser.Dependency{},
	}
	r := NewDependencyResolver(analysis)

	// Test suggestions
	plan, err := r.ResolveDependencies([]string{"user"})
	assert.NoError(t, err)

	// user suggests notification and audit
	assert.Contains(t, plan.OptionalModules, "audit")
	assert.Contains(t, plan.OptionalModules, "notification")
}
