package depanalysis

import (
	"fmt"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Analyze inspects internal packages under the given root and enforces layering
// rules while also detecting import cycles.
func Analyze(root string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedImports,
		Dir:  root,
	}
	patterns := []string{
		"./internal/adapters/...",
		"./internal/usecase/...",
		"./internal/repository/...",
		"./internal/infrastructure/...",
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			if strings.Contains(e.Msg, "import cycle") {
				if idx := strings.Index(e.Msg, "import stack:"); idx >= 0 {
					stack := strings.Trim(e.Msg[idx+len("import stack:"):], " []")
					stack = strings.ReplaceAll(stack, " ", " -> ")
					return fmt.Errorf("import cycle detected: %s", stack)
				}
				return fmt.Errorf("import cycle detected: %s", e.Msg)
			}
		}
	}
	layerIndex := map[string]int{}
	for _, pkg := range pkgs {
		switch {
		case strings.Contains(pkg.PkgPath, "/internal/adapters/"):
			layerIndex[pkg.PkgPath] = 0
		case strings.Contains(pkg.PkgPath, "/internal/usecase/"):
			layerIndex[pkg.PkgPath] = 1
		case strings.Contains(pkg.PkgPath, "/internal/repository/"):
			layerIndex[pkg.PkgPath] = 2
		case strings.Contains(pkg.PkgPath, "/internal/infrastructure/"):
			layerIndex[pkg.PkgPath] = 3
		}
	}

	graph := make(map[string][]string)
	for _, pkg := range pkgs {
		srcIdx, ok := layerIndex[pkg.PkgPath]
		if !ok {
			continue
		}
		for _, imp := range pkg.Imports {
			dstIdx, ok := layerIndex[imp.PkgPath]
			if !ok {
				continue
			}
			graph[pkg.PkgPath] = append(graph[pkg.PkgPath], imp.PkgPath)
			if dstIdx < srcIdx {
				return fmt.Errorf("layer violation: %s (layer %d) imports %s (layer %d)", pkg.PkgPath, srcIdx, imp.PkgPath, dstIdx)
			}
		}
	}

	visited := make(map[string]bool)
	stack := make(map[string]bool)
	var path []string
	var dfs func(string) error
	dfs = func(n string) error {
		if stack[n] {
			cycle := append([]string{}, path...)
			cycle = append(cycle, n)
			idx := indexOf(cycle, n)
			cycle = cycle[idx:]
			return fmt.Errorf("import cycle detected: %s", strings.Join(cycle, " -> "))
		}
		if visited[n] {
			return nil
		}
		visited[n] = true
		stack[n] = true
		path = append(path, n)
		for _, m := range graph[n] {
			if err := dfs(m); err != nil {
				return err
			}
		}
		path = path[:len(path)-1]
		delete(stack, n)
		return nil
	}
	for node := range graph {
		if err := dfs(node); err != nil {
			return err
		}
	}
	return nil
}

func indexOf(s []string, target string) int {
	for i, v := range s {
		if v == target {
			return i
		}
	}
	return -1
}
