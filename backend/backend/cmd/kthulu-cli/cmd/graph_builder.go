package cmd

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/graph"
)

// BuildValidationGraph scans use cases and adapters to construct a graph
// representing relationships between modules, use cases and adapters. It is
// intended for validation and visualization purposes.
func BuildValidationGraph(config *core.Config) (*graph.Graph, error) {
	g := graph.New()

	// Map of usecase name -> module for quick lookup when processing adapters
	usecaseModules := make(map[string]string)

	usecaseDir := "github.com/pmaojo/kthulu-go/backend/internal/usecase"

	// First, add module and usecase nodes
	for _, module := range config.Modules {
		moduleNode := fmt.Sprintf("module:%s", module)
		g.AddNode(moduleNode)

		// Scan usecase directory for files tagged with module annotation
		err := filepath.WalkDir(usecaseDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			tag := fmt.Sprintf("// @kthulu:module:%s", module)
			if !strings.Contains(string(data), tag) {
				return nil
			}
			usecaseName := strings.TrimSuffix(filepath.Base(path), ".go")
			usecaseNode := fmt.Sprintf("usecase:%s", usecaseName)
			g.AddNode(usecaseNode)
			if err := g.AddEdge(moduleNode, usecaseNode); err != nil {
				return err
			}
			usecaseModules[usecaseName] = module
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Prepare list of known usecase names for adapter scanning
	usecaseNames := make([]string, 0, len(usecaseModules))
	for name := range usecaseModules {
		usecaseNames = append(usecaseNames, name)
	}

	adaptersDir := "github.com/pmaojo/kthulu-go/backend/internal/adapters"

	// Now scan adapters and build edges
	for _, module := range config.Modules {
		err := filepath.WalkDir(adaptersDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			tag := fmt.Sprintf("// @kthulu:module:%s", module)
			if !strings.Contains(string(data), tag) {
				return nil
			}

			adapterName := strings.TrimSuffix(filepath.Base(path), ".go")
			adapterNode := fmt.Sprintf("adapter:%s", adapterName)
			g.AddNode(adapterNode)

			moduleNode := fmt.Sprintf("module:%s", module)
			if err := g.AddEdge(adapterNode, moduleNode); err != nil {
				return err
			}

			// Parse AST to detect references to usecases
			fset := token.NewFileSet()
			fileAST, err := parser.ParseFile(fset, path, data, parser.AllErrors)
			if err != nil {
				return nil
			}

			// Collect aliases for imports referencing the usecase package
			aliases := map[string]struct{}{}
			for _, imp := range fileAST.Imports {
				impPath, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					continue
				}
				if strings.Contains(impPath, "internal/usecase") {
					alias := "usecase"
					if imp.Name != nil {
						alias = imp.Name.Name
					}
					aliases[alias] = struct{}{}
				}
			}
			if len(aliases) == 0 {
				return nil
			}

			// Traverse AST to find selector expressions referencing the alias
			ast.Inspect(fileAST, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if _, ok := aliases[ident.Name]; !ok {
					return true
				}
				selLower := strings.ToLower(sel.Sel.Name)
				for _, uc := range usecaseNames {
					if strings.Contains(selLower, strings.ToLower(uc)) {
						ucNode := fmt.Sprintf("usecase:%s", uc)
						_ = g.AddEdge(adapterNode, ucNode)
					}
				}
				return true
			})

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return g, nil
}
