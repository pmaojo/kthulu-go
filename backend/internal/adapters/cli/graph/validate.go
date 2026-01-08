package graph

import (
	"fmt"
	"strings"
)

// ValidateGraph checks the provided graph against a set of rules:
//   - Adapters must not depend directly on repositories.
//   - Every usecase must be triggered by at least one adapter.
//   - The overall graph must be acyclic.
//
// It returns an error containing all violations found, or nil if the graph satisfies the rules.
func ValidateGraph(g *Graph) error {
	if g == nil {
		return fmt.Errorf("graph is nil")
	}

	var violations []string

	// Rule 1: Detect edges from adapter to repository
	for from, edges := range g.Edges {
		if isAdapter(from) {
			for _, e := range edges {
				if isRepository(e.To) {
					violations = append(violations, fmt.Sprintf("adapter %s depends on repository %s", from, e.To))
				}
			}
		}
	}

	// Rule 2: Each usecase must have an incoming edge from an adapter
	for id := range g.Nodes {
		if !isUsecase(id) {
			continue
		}
		hasIncoming := false
		for from, edges := range g.Edges {
			if !isAdapter(from) {
				continue
			}
			for _, e := range edges {
				if e.To == id {
					hasIncoming = true
					break
				}
			}
			if hasIncoming {
				break
			}
		}
		if !hasIncoming {
			violations = append(violations, fmt.Sprintf("usecase %s has no incoming edge from adapter", id))
		}
	}

	// Rule 3: Detect cycles in the graph
	cycles := findCycles(g)
	for _, c := range cycles {
		violations = append(violations, fmt.Sprintf("cycle detected: %s", strings.Join(c, " -> ")))
	}

	if len(violations) > 0 {
		return fmt.Errorf("%s", strings.Join(violations, "; "))
	}
	return nil
}

func isAdapter(id string) bool {
	return strings.Contains(id, "adapter")
}

func isRepository(id string) bool {
	return strings.Contains(id, "repository")
}

func isUsecase(id string) bool {
	return strings.Contains(id, "usecase")
}

// findCycles returns a slice of cycles, where each cycle is represented
// as an ordered list of node IDs forming that cycle.
func findCycles(g *Graph) [][]string {
	var cycles [][]string
	state := make(map[string]int) // 0=unvisited,1=visiting,2=visited
	var stack []string
	seen := make(map[string]struct{})

	var dfs func(string)
	dfs = func(node string) {
		state[node] = 1
		stack = append(stack, node)
		for _, e := range g.Edges[node] {
			to := e.To
			if state[to] == 0 {
				dfs(to)
			} else if state[to] == 1 {
				// cycle detected
				idx := -1
				for i, n := range stack {
					if n == to {
						idx = i
						break
					}
				}
				if idx >= 0 {
					cycle := append([]string{}, stack[idx:]...)
					cycle = append(cycle, to)
					key := strings.Join(cycle, "->")
					if _, ok := seen[key]; !ok {
						cycles = append(cycles, cycle)
						seen[key] = struct{}{}
					}
				}
			}
		}
		stack = stack[:len(stack)-1]
		state[node] = 2
	}

	for id := range g.Nodes {
		if state[id] == 0 {
			dfs(id)
		}
	}
	return cycles
}
