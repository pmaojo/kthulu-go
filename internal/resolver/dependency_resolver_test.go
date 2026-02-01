package resolver

import (
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/stretchr/testify/assert"
)

func TestDependencyResolver_GetModuleInfo(t *testing.T) {
	modules := map[string]*parser.Module{
		"user": {
			Name:    "user",
			Package: "github.com/example/user",
			Files:   []string{"user.go"},
		},
		"unknown": {
			Name:    "unknown",
			Package: "github.com/example/unknown",
			Files:   []string{"unknown.go"},
		},
	}

	analysis := &parser.ProjectAnalysis{
		Modules:      modules,
		Dependencies: []parser.Dependency{},
	}

	resolver := NewDependencyResolver(analysis)

	t.Run("Known Module", func(t *testing.T) {
		info, err := resolver.GetModuleInfo("user")
		assert.NoError(t, err)
		assert.Equal(t, "User management and authentication core", info.Description)
		assert.Equal(t, "Core", info.Category)
	})

	t.Run("Custom Module", func(t *testing.T) {
		info, err := resolver.GetModuleInfo("unknown")
		assert.NoError(t, err)
		assert.Equal(t, "Custom module", info.Description)
		assert.Equal(t, "Custom", info.Category)
	})
}

func TestDependencyResolver_CheckIncompatibleModules(t *testing.T) {
	// Setup resolver
	analysis := &parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}
	resolver := NewDependencyResolver(analysis)

	// Create a plan with incompatible modules
	plan := &ResolutionPlan{
		RequiredModules: []string{"mysql", "postgresql"},
		Conflicts:       []ConflictInfo{},
	}

	// Manually trigger conflict detection (since it's private, we test the public ResolveDependencies or use reflection/internal test)
	// Since we are in the same package `resolver`, we can access private methods if the test is in `resolver` package (not `resolver_test`)
	// The file declared `package resolver`, so we can access `checkIncompatibleModules`.

	resolver.checkIncompatibleModules(plan)

	assert.NotEmpty(t, plan.Conflicts)
	assert.Equal(t, "incompatible", plan.Conflicts[0].Type)
	assert.Contains(t, plan.Conflicts[0].Modules, "mysql")
	assert.Contains(t, plan.Conflicts[0].Modules, "postgresql")
}
