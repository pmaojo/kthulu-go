package graph

import (
	"strings"
	"testing"
)

func TestValidateGraph_AdapterToRepositoryViolation(t *testing.T) {
	g := New()
	g.AddNode("adapter1")
	g.AddNode("repository1")
	_ = g.AddEdge("adapter1", "repository1")

	err := ValidateGraph(g)
	if err == nil || !strings.Contains(err.Error(), "adapter adapter1 depends on repository repository1") {
		t.Fatalf("expected adapter->repository violation, got %v", err)
	}
}

func TestValidateGraph_UsecaseRequiresAdapter(t *testing.T) {
	g := New()
	g.AddNode("adapter1")
	g.AddNode("usecase1")
	// no edge from adapter1 to usecase1

	err := ValidateGraph(g)
	if err == nil || !strings.Contains(err.Error(), "usecase usecase1 has no incoming edge from adapter") {
		t.Fatalf("expected missing adapter->usecase violation, got %v", err)
	}
}

func TestValidateGraph_CycleDetection(t *testing.T) {
	g := New()
	g.AddNode("A")
	g.AddNode("B")
	g.AddNode("C")
	_ = g.AddEdge("A", "B")
	_ = g.AddEdge("B", "C")
	_ = g.AddEdge("C", "A")

	err := ValidateGraph(g)
	if err == nil || !strings.Contains(err.Error(), "cycle detected") {
		t.Fatalf("expected cycle detection error, got %v", err)
	}
}

func TestValidateGraph_NoViolations(t *testing.T) {
	g := New()
	g.AddNode("adapter1")
	g.AddNode("usecase1")
	g.AddNode("repository1")
	_ = g.AddEdge("adapter1", "usecase1")
	_ = g.AddEdge("usecase1", "repository1")

	if err := ValidateGraph(g); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
