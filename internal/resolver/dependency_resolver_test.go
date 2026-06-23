package resolver

import (
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/stretchr/testify/assert"
)

func TestResolveDependencies(t *testing.T) {
	// Setup mock analysis
	analysis := &parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}

	resolver := NewDependencyResolver(analysis)

	// Test resolving modules with dependencies
	// auth depends on user
	plan, err := resolver.ResolveDependencies([]string{"auth"})
	assert.NoError(t, err)
	assert.Contains(t, plan.RequiredModules, "auth")
	assert.Contains(t, plan.RequiredModules, "user")
	assert.Len(t, plan.RequiredModules, 2)

	// Test resolving modules with multiple dependencies
	// organization depends on user, auth
	plan, err = resolver.ResolveDependencies([]string{"organization"})
	assert.NoError(t, err)
	assert.Contains(t, plan.RequiredModules, "organization")
	assert.Contains(t, plan.RequiredModules, "user")
	assert.Contains(t, plan.RequiredModules, "auth")
	assert.Len(t, plan.RequiredModules, 3)
}

func TestGetModuleInfo(t *testing.T) {
	// Setup mock analysis with a module
	analysis := &parser.ProjectAnalysis{
		Modules: map[string]*parser.Module{
			"user": {
				Name:         "user",
				Package:      "github.com/pmaojo/kthulu-go/internal/modules/user",
				Dependencies: []string{},
				Files:        []string{"user.go"},
			},
		},
		Dependencies: []parser.Dependency{},
	}

	resolver := NewDependencyResolver(analysis)

	// Test getting module info
	info, err := resolver.GetModuleInfo("user")
	assert.NoError(t, err)
	assert.Equal(t, "user", info.Name)
	assert.Equal(t, "User management and authentication core", info.Description)
	assert.Equal(t, "Core", info.Category)

	// Test unknown module
	_, err = resolver.GetModuleInfo("unknown")
	assert.Error(t, err)
}

func TestConflictDetection(t *testing.T) {
	analysis := &parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}
	resolver := NewDependencyResolver(analysis)

	// Test incompatible modules
	// sqlite and postgresql are incompatible
	plan, err := resolver.ResolveDependencies([]string{"sqlite", "postgresql"})
	assert.NoError(t, err) // It resolves, but reports conflicts
	assert.NotEmpty(t, plan.Conflicts)

	found := false
	for _, c := range plan.Conflicts {
		if c.Type == "incompatible" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should detect incompatible modules")
}
