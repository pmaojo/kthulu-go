package resolver

import (
	"testing"
	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
)

func BenchmarkGetModuleDescription(b *testing.B) {
	r := NewDependencyResolver(&parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.getModuleDescription("user")
	}
}

func BenchmarkGetModuleCategory(b *testing.B) {
	r := NewDependencyResolver(&parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.getModuleCategory("user")
	}
}
