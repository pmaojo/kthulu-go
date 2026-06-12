package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// RoutesListArgs for routes_list tool.
type RoutesListArgs struct {
	Module string `json:"module,omitempty" jsonschema:"description=Filter routes by module name (e.g. 'order' shows only /orders/* routes)."`
	Format string `json:"format,omitempty" jsonschema:"description=Output format: 'table' (default) or 'json'."`
}

// RouteEntry represents a single discovered HTTP route.
type RouteEntry struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	Handler    string `json:"handler"`
	Module     string `json:"module"`
	Middleware string `json:"middleware,omitempty"`
}

// routesListTool builds the routes_list RegisteredTool.
func routesListTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name: "routes_list",
		Description: "List all HTTP routes registered in this project. Shows method, path, handler name, and detected middleware. " +
			"Essential for understanding the current API surface before adding endpoints.",
		Handler: func(ctx context.Context, args RoutesListArgs) (*mcp_golang.ToolResponse, error) {
			dir := resolveWorkdir(workingDir)

			routes, err := DiscoverRoutes(dir)
			if err != nil {
				return nil, fmt.Errorf("route discovery failed: %w", err)
			}

			// Apply module filter.
			if args.Module != "" {
				routes = FilterRoutesByModule(routes, args.Module)
			}

			if len(routes) == 0 {
				msg := "No HTTP routes found"
				if args.Module != "" {
					msg += fmt.Sprintf(" matching module %q", args.Module)
				}
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(msg)), nil
			}

			// Output format.
			format := strings.ToLower(strings.TrimSpace(args.Format))
			if format == "json" {
				out, err := FormatRoutesJSON(routes)
				if err != nil {
					return nil, err
				}
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(out)), nil
			}

			// Default: table.
			out := FormatRoutesTable(routes)
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(truncateOutput(out))), nil
		},
	}
}

// DiscoverRoutes walks all .go files under dir, uses AST to find route
// registrations, and falls back to grep when AST finds nothing.
func DiscoverRoutes(dir string) ([]RouteEntry, error) {
	var routes []RouteEntry

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}

		found, parseErr := parseFileRoutes(dir, path)
		if parseErr == nil {
			routes = append(routes, found...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Fall back to grep when AST finds nothing.
	if len(routes) == 0 {
		routes = discoverRoutesGrep(dir)
	}
	return routes, nil
}

// FilterRoutesByModule filters routes whose module or path contains the given module string.
func FilterRoutesByModule(routes []RouteEntry, module string) []RouteEntry {
	if module == "" {
		return routes
	}
	mod := strings.ToLower(module)
	var filtered []RouteEntry
	for _, r := range routes {
		if strings.Contains(strings.ToLower(r.Module), mod) ||
			strings.Contains(strings.ToLower(r.Path), mod) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// FormatRoutesTable renders routes as a padded text table.
func FormatRoutesTable(routes []RouteEntry) string {
	const fmtStr = "%-8s %-32s %-30s %-12s %s\n"
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(fmtStr, "METHOD", "PATH", "HANDLER", "MODULE", "MIDDLEWARE"))
	sb.WriteString(strings.Repeat("-", 100) + "\n")
	for _, r := range routes {
		sb.WriteString(fmt.Sprintf(fmtStr, r.Method, r.Path, r.Handler, r.Module, r.Middleware))
	}
	return sb.String()
}

// FormatRoutesJSON serialises routes as a JSON array string.
func FormatRoutesJSON(routes []RouteEntry) (string, error) {
	data, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// routeMethodMap maps well-known Go method selector names to their HTTP verb.
var routeMethodMap = map[string]string{
	"GET":     "GET",
	"POST":    "POST",
	"PUT":     "PUT",
	"PATCH":   "PATCH",
	"DELETE":  "DELETE",
	"Get":     "GET",
	"Post":    "POST",
	"Put":     "PUT",
	"Patch":   "PATCH",
	"Delete":  "DELETE",
	"Head":    "HEAD",
	"Options": "OPTIONS",
}

// parseFileRoutes parses a single Go file and extracts HTTP route registrations.
func parseFileRoutes(workingDir, filePath string) ([]RouteEntry, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	rel, relErr := filepath.Rel(workingDir, filePath)
	if relErr != nil {
		rel = filePath
	}
	module := inferModule(rel)

	var routes []RouteEntry

	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		methodName := sel.Sel.Name

		// Pattern: r.GET("/path", handler) / r.Get("/path", handler) etc.
		if httpMethod, known := routeMethodMap[methodName]; known {
			if len(call.Args) >= 2 {
				path := stringLiteralVal(call.Args[0])
				if path != "" {
					handler, middleware := handlerInfo(call.Args[len(call.Args)-1])
					routes = append(routes, RouteEntry{
						Method:     httpMethod,
						Path:       path,
						Handler:    handler,
						Module:     module,
						Middleware: middleware,
					})
					return true
				}
			}
		}

		// Pattern: r.Handle("METHOD", "/path", handler)
		if methodName == "Handle" && len(call.Args) >= 3 {
			method := strings.ToUpper(stringLiteralVal(call.Args[0]))
			path := stringLiteralVal(call.Args[1])
			if method != "" && path != "" {
				handler, middleware := handlerInfo(call.Args[len(call.Args)-1])
				routes = append(routes, RouteEntry{
					Method:     method,
					Path:       path,
					Handler:    handler,
					Module:     module,
					Middleware: middleware,
				})
				return true
			}
		}

		// Pattern: router.HandleFunc("/path", handler).Methods("GET")
		if methodName == "Methods" && len(call.Args) >= 1 {
			methods := collectStringArgs(call.Args)
			if inner, ok := sel.X.(*ast.CallExpr); ok {
				if innerSel, ok := inner.Fun.(*ast.SelectorExpr); ok {
					if innerSel.Sel.Name == "HandleFunc" && len(inner.Args) >= 2 {
						path := stringLiteralVal(inner.Args[0])
						if path != "" {
							handler, middleware := handlerInfo(inner.Args[len(inner.Args)-1])
							for _, m := range methods {
								routes = append(routes, RouteEntry{
									Method:     strings.ToUpper(m),
									Path:       path,
									Handler:    handler,
									Module:     module,
									Middleware: middleware,
								})
							}
							return true
						}
					}
				}
			}
		}

		return true
	})

	return routes, nil
}

// stringLiteralVal returns the string value of a basic literal, or "".
func stringLiteralVal(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	s := lit.Value
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		return s[1 : len(s)-1]
	}
	return ""
}

// collectStringArgs returns the string value of each basic-literal argument.
func collectStringArgs(args []ast.Expr) []string {
	var out []string
	for _, a := range args {
		if s := stringLiteralVal(a); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// handlerInfo returns the handler name and any detected middleware from an expression.
// When the arg is a call like authMiddleware(handler), the outer call name is
// returned as middleware and the inner expression as the handler.
func handlerInfo(expr ast.Expr) (handler, middleware string) {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name, ""
	case *ast.SelectorExpr:
		if x, ok := e.X.(*ast.Ident); ok {
			return x.Name + "." + e.Sel.Name, ""
		}
		return e.Sel.Name, ""
	case *ast.CallExpr:
		mw := callExprName(e.Fun)
		if len(e.Args) > 0 {
			inner, _ := handlerInfo(e.Args[len(e.Args)-1])
			return inner, mw
		}
		return mw, ""
	}
	return "?", ""
}

// callExprName returns a human-readable name for a function expression.
func callExprName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		if x, ok := e.X.(*ast.Ident); ok {
			return x.Name + "." + e.Sel.Name
		}
		return e.Sel.Name
	}
	return "?"
}

// inferModule derives a module name from a relative file path.
// e.g. "internal/order/router.go" → "order"
func inferModule(rel string) string {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for i := len(parts) - 2; i >= 0; i-- {
		p := parts[i]
		switch p {
		case "internal", "cmd", "server", "router", "handlers", "http", "api", "adapters":
			continue
		}
		return p
	}
	return ""
}

// discoverRoutesGrep falls back to grep when AST finds nothing.
func discoverRoutesGrep(dir string) []RouteEntry {
	pattern := `r\.GET\|r\.POST\|r\.PUT\|r\.PATCH\|r\.DELETE\|r\.Get\|r\.Post\|r\.Put\|r\.Patch\|r\.Delete`
	cmd := exec.Command("grep", "-rn", pattern, "--include=*.go", dir)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var routes []RouteEntry
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: /path/to/file.go:123:  r.GET("/foo", handler)
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}
		filePart := line[:colonIdx]
		rest := line[colonIdx+1:]
		colonIdx2 := strings.Index(rest, ":")
		if colonIdx2 < 0 {
			continue
		}
		codePart := strings.TrimSpace(rest[colonIdx2+1:])

		module := inferModule(filePart)

		method := ""
		for _, m := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "Get", "Post", "Put", "Patch", "Delete"} {
			if strings.Contains(codePart, "."+m+"(") {
				method = strings.ToUpper(m)
				break
			}
		}
		if method == "" {
			continue
		}

		path := ""
		if qi := strings.Index(codePart, `"`); qi >= 0 {
			rest2 := codePart[qi+1:]
			if qi2 := strings.Index(rest2, `"`); qi2 >= 0 {
				path = rest2[:qi2]
			}
		}
		if path == "" {
			continue
		}

		routes = append(routes, RouteEntry{
			Method:  method,
			Path:    path,
			Handler: "?",
			Module:  module,
		})
	}
	return routes
}

func init() {
	RegisterPlugin(func(_ CommandExecutor, workingDir string) []RegisteredTool {
		return []RegisteredTool{routesListTool(workingDir)}
	})
}
