package graph

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAddNode(t *testing.T) {
	g := New()
	g.AddNode("A")
	if _, ok := g.Nodes["A"]; !ok {
		t.Fatalf("node A not added")
	}
}

func TestAddEdgeAndIncoming(t *testing.T) {
	g := New()
	g.AddNode("A")
	g.AddNode("B")
	if err := g.AddEdge("A", "B"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Edges["A"]) != 1 || g.Edges["A"][0].To != "B" {
		t.Fatalf("edge from A to B not stored")
	}
	if !g.HasIncomingEdges("B") {
		t.Fatalf("expected B to have incoming edge")
	}
	if g.HasIncomingEdges("A") {
		t.Fatalf("did not expect A to have incoming edge")
	}
}

func TestSerialization(t *testing.T) {
	g := New()
	g.AddNode("A")
	g.AddNode("B")
	_ = g.AddEdge("A", "B")

	dot := g.ToDOT()
	if !strings.Contains(dot, "\"A\" -> \"B\"") {
		t.Fatalf("DOT serialization missing edge: %s", dot)
	}

	data, err := g.ToJSON()
	if err != nil {
		t.Fatalf("json serialization error: %v", err)
	}
	var out Graph
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if len(out.Nodes) != 2 {
		t.Fatalf("expected 2 nodes in json, got %d", len(out.Nodes))
	}
}
