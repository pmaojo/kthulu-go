package graph

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Node represents a graph node.
type Node struct {
	ID string `json:"id"`
}

// Edge represents a directed edge between two nodes.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Graph is a simple directed graph.
type Graph struct {
	Nodes map[string]*Node   `json:"nodes"`
	Edges map[string][]*Edge `json:"edges"`
}

// New creates a new Graph instance.
func New() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]*Edge),
	}
}

// AddNode inserts a node into the graph.
func (g *Graph) AddNode(id string) {
	if g.Nodes == nil {
		g.Nodes = make(map[string]*Node)
	}
	if _, exists := g.Nodes[id]; !exists {
		g.Nodes[id] = &Node{ID: id}
	}
}

// AddEdge creates a directed edge from one node to another.
func (g *Graph) AddEdge(from, to string) error {
	if g.Nodes == nil {
		g.Nodes = make(map[string]*Node)
	}
	if g.Edges == nil {
		g.Edges = make(map[string][]*Edge)
	}
	if _, ok := g.Nodes[from]; !ok {
		return fmt.Errorf("from node %s not found", from)
	}
	if _, ok := g.Nodes[to]; !ok {
		return fmt.Errorf("to node %s not found", to)
	}
	edge := &Edge{From: from, To: to}
	g.Edges[from] = append(g.Edges[from], edge)
	return nil
}

// HasIncomingEdges reports whether the node has any incoming edges.
func (g *Graph) HasIncomingEdges(id string) bool {
	for _, edges := range g.Edges {
		for _, e := range edges {
			if e.To == id {
				return true
			}
		}
	}
	return false
}

// ToDOT exports the graph in Graphviz DOT format.
func (g *Graph) ToDOT() string {
	var sb strings.Builder
	sb.WriteString("digraph {\n")
	for id := range g.Nodes {
		sb.WriteString(fmt.Sprintf("  \"%s\";\n", id))
	}
	for from, edges := range g.Edges {
		for _, e := range edges {
			sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", from, e.To))
		}
	}
	sb.WriteString("}\n")
	return sb.String()
}

// ToJSON serializes the graph to JSON.
func (g *Graph) ToJSON() ([]byte, error) {
	return json.Marshal(g)
}
