package resolver

import (
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/stretchr/testify/assert"
)

func TestGetModuleInfo(t *testing.T) {
	// Setup mock analysis
	mockModules := map[string]*parser.Module{
		"user": {
			Name:         "user",
			Package:      "user",
			Files:        []string{"user.go"},
			Dependencies: []string{},
		},
		"custom": {
			Name:         "custom",
			Package:      "custom",
			Files:        []string{"custom.go"},
			Dependencies: []string{"user"},
		},
	}
	analysis := &parser.ProjectAnalysis{
		Modules:      mockModules,
		Dependencies: []parser.Dependency{},
	}

	r := NewDependencyResolver(analysis)

	// Test known module
	info, err := r.GetModuleInfo("user")
	assert.NoError(t, err)
	assert.Equal(t, "user", info.Name)
	assert.Equal(t, "User management and authentication core", info.Description)
	assert.Equal(t, "Core", info.Category)

	// Test custom module
	info, err = r.GetModuleInfo("custom")
	assert.NoError(t, err)
	assert.Equal(t, "custom", info.Name)
	assert.Equal(t, "Custom module", info.Description)
	assert.Equal(t, "Custom", info.Category)

	// Test missing module
	_, err = r.GetModuleInfo("missing")
	assert.Error(t, err)
}

func TestResolveDependencies_ConflictsAndOptionals(t *testing.T) {
	analysis := &parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}
	r := NewDependencyResolver(analysis)

	// Test incompatible modules
	// Note: mysql and postgresql are not in r.rules by default, so they are treated as modules with no deps.
	plan, err := r.ResolveDependencies([]string{"mysql", "postgresql"})
	assert.NoError(t, err)

	hasConflict := false
	for _, c := range plan.Conflicts {
		if c.Type == "incompatible" {
			hasConflict = true
			break
		}
	}
	assert.True(t, hasConflict, "Should detect incompatible modules (mysql vs postgresql)")

	// Test optional suggestions
	// "user" suggests "notification", "audit"
	plan, err = r.ResolveDependencies([]string{"user"})
	assert.NoError(t, err)
	assert.Contains(t, plan.OptionalModules, "audit")
	assert.Contains(t, plan.OptionalModules, "notification")
}
