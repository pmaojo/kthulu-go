package resolver

import (
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/stretchr/testify/assert"
)

func TestDependencyResolver_ResolveDependencies(t *testing.T) {
	// Setup mock analysis
	analysis := &parser.ProjectAnalysis{
		Modules: map[string]*parser.Module{
			"user": {Name: "user", Package: "user"},
			"auth": {Name: "auth", Package: "auth"},
			"audit": {Name: "audit", Package: "audit"},
			"payment": {Name: "payment", Package: "payment"},
			"invoice": {Name: "invoice", Package: "invoice"},
			"sqlite": {Name: "sqlite", Package: "sqlite"},
			"mysql": {Name: "mysql", Package: "mysql"},
			"analytics": {Name: "analytics", Package: "analytics"},
		},
		Dependencies: []parser.Dependency{},
	}

	resolver := NewDependencyResolver(analysis)

	t.Run("Resolves dependencies correctly", func(t *testing.T) {
		plan, err := resolver.ResolveDependencies([]string{"payment"})
		assert.NoError(t, err)
		assert.Contains(t, plan.RequiredModules, "payment")
		// payment -> invoice -> user, organization...
		assert.Contains(t, plan.RequiredModules, "user")
		assert.Contains(t, plan.RequiredModules, "invoice")
	})

	t.Run("Detects conflicts", func(t *testing.T) {
		// sqlite and mysql are incompatible
		plan, err := resolver.ResolveDependencies([]string{"sqlite", "mysql"})
		assert.NoError(t, err)

		var foundConflict bool
		for _, c := range plan.Conflicts {
			if c.Type == "incompatible" {
				foundConflict = true
				break
			}
		}
		assert.True(t, foundConflict, "Should detect incompatible sqlite and mysql")
	})

	t.Run("Suggests optional modules", func(t *testing.T) {
		plan, err := resolver.ResolveDependencies([]string{"user"})
		assert.NoError(t, err)
		// user suggests notification and audit
		assert.Contains(t, plan.OptionalModules, "audit")
	})

	t.Run("Generates recommendations", func(t *testing.T) {
		// analytics + audit -> growth-engine recommendation
		plan, err := resolver.ResolveDependencies([]string{"analytics", "audit"})
		assert.NoError(t, err)

		var foundRec bool
		for _, r := range plan.Recommendations {
			if r.Module == "growth-engine" && r.Type == "wire" {
				foundRec = true
				break
			}
		}
		assert.True(t, foundRec, "Should recommend growth-engine")
	})
}

func TestDependencyResolver_GetModuleInfo(t *testing.T) {
	analysis := &parser.ProjectAnalysis{
		Modules: map[string]*parser.Module{
			"user": {Name: "user", Package: "user", Files: []string{"a.go"}, Dependencies: []string{}},
		},
	}
	resolver := NewDependencyResolver(analysis)

	info, err := resolver.GetModuleInfo("user")
	assert.NoError(t, err)
	assert.Equal(t, "user", info.Name)
	assert.Equal(t, "User management and authentication core", info.Description)
	assert.Equal(t, "Core", info.Category)
}
