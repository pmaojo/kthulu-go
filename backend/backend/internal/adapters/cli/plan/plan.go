package plan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// ActionType represents how a construct should be applied.
type ActionType string

const (
	// Replace fully substitutes the target path with the construct.
	Replace ActionType = "Replace"
	// Decorate applies additional changes on top of an existing construct.
	Decorate ActionType = "Decorate"
)

// Construct is a unit produced by the scanner that targets a specific path.
type Construct struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Priority int    `json:"priority"`
}

// Node is an element in the plan DAG. Each node wraps a construct and the
// action that should be applied.
type Node struct {
	Construct Construct  `json:"construct"`
	Action    ActionType `json:"action"`
}

// Edge represents a parent-child relationship between two constructs in the DAG.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Plan is the result of analyzing scanner output.
type Plan struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Build constructs a deterministic plan from a slice of constructs.
func Build(constructs []Construct) *Plan {
	// group constructs by path
	byPath := make(map[string][]Construct)
	for _, c := range constructs {
		byPath[c.Path] = append(byPath[c.Path], c)
	}

	var nodes []Node
	for path, group := range byPath {
		// sort by priority desc, then ID asc for determinism
		sort.Slice(group, func(i, j int) bool {
			if group[i].Priority == group[j].Priority {
				return group[i].ID < group[j].ID
			}
			return group[i].Priority > group[j].Priority
		})
		for i, c := range group {
			action := Decorate
			if i == 0 {
				action = Replace
			}
			nodes = append(nodes, Node{Construct: c, Action: action})
		}
		_ = path
	}

	// build edges based on parent directories present in constructs
	pathSet := make(map[string]struct{})
	for _, n := range nodes {
		pathSet[n.Construct.Path] = struct{}{}
	}
	var edges []Edge
	edgeSet := make(map[Edge]struct{})
	for _, n := range nodes {
		parent := filepath.Dir(n.Construct.Path)
		if parent == "." || parent == n.Construct.Path {
			continue
		}
		e := Edge{From: parent, To: n.Construct.Path}
		if _, ok := pathSet[parent]; ok {
			if _, exists := edgeSet[e]; !exists {
				edgeSet[e] = struct{}{}
				edges = append(edges, e)
			}
		}
	}

	// deterministic ordering
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Construct.Path == nodes[j].Construct.Path {
			if nodes[i].Action == nodes[j].Action {
				return nodes[i].Construct.ID < nodes[j].Construct.ID
			}
			return nodes[i].Action < nodes[j].Action
		}
		return nodes[i].Construct.Path < nodes[j].Construct.Path
	})
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From == edges[j].From {
			return edges[i].To < edges[j].To
		}
		return edges[i].From < edges[j].From
	})

	return &Plan{Nodes: nodes, Edges: edges}
}

// Write stores the plan as JSON at <root>/.kthulu/plan.json.
func Write(p *Plan, root string) error {
	outDir := filepath.Join(root, ".kthulu")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	outPath := filepath.Join(outDir, "plan.json")
	return os.WriteFile(outPath, data, 0o644)
}
