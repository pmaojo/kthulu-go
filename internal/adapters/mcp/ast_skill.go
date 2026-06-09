package mcp

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// GoASTService exposes Go source analysis tools backed by go/parser: file and
// package outlines, project-wide symbol lookup, and exact symbol source
// extraction.
type GoASTService struct{}

// NewGoASTService creates a new GoASTService.
func NewGoASTService() *GoASTService {
	return &GoASTService{}
}

// GetTools returns all AST tools bound to the working directory.
func (s *GoASTService) GetTools(workingDir string) []RegisteredTool {
	return []RegisteredTool{
		s.outlineTool(workingDir),
		s.findSymbolTool(workingDir),
		s.symbolSourceTool(workingDir),
	}
}

// GoOutlineArgs defines arguments for outlining a file or package.
type GoOutlineArgs struct {
	Path         string `json:"path" jsonschema:"description=Go file or package directory to outline (relative to the project root),required"`
	IncludeTests bool   `json:"include_tests,omitempty" jsonschema:"description=Include _test.go files when outlining a directory"`
}

// GoFindSymbolArgs defines arguments for project-wide symbol lookup.
type GoFindSymbolArgs struct {
	Name       string `json:"name" jsonschema:"description=Symbol name to find (function, method, type, const, or var),required"`
	Exact      bool   `json:"exact,omitempty" jsonschema:"description=Require an exact name match instead of a case-insensitive substring match"`
	MaxResults int    `json:"max_results,omitempty" jsonschema:"description=Maximum number of symbols to return (default 50)"`
}

// GoSymbolSourceArgs defines arguments for extracting a symbol's source code.
type GoSymbolSourceArgs struct {
	Name string `json:"name" jsonschema:"description=Exact name of the symbol whose source to extract,required"`
	Path string `json:"path,omitempty" jsonschema:"description=Go file to look in; omit to search the whole project"`
}

// goSymbol is a single declaration found while parsing.
type goSymbol struct {
	Kind      string
	Name      string
	Signature string
	File      string
	Line      int
	EndLine   int
	Doc       string
}

func (s *GoASTService) outlineTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "go_outline",
		Description: "Parse a Go file or package directory and list its symbols (functions, methods, types, consts, vars) with signatures and line numbers.",
		Handler: func(ctx context.Context, args GoOutlineArgs) (*mcp_golang.ToolResponse, error) {
			outline, err := s.Outline(workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(outline)), nil
		},
	}
}

func (s *GoASTService) findSymbolTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "go_find_symbol",
		Description: "Find Go symbols by name across the whole project. Returns kind, signature, and file:line for each match.",
		Handler: func(ctx context.Context, args GoFindSymbolArgs) (*mcp_golang.ToolResponse, error) {
			result, err := s.FindSymbol(workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}

func (s *GoASTService) symbolSourceTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "go_symbol_source",
		Description: "Extract the exact source code (including doc comment) of a named Go symbol from a file or the whole project.",
		Handler: func(ctx context.Context, args GoSymbolSourceArgs) (*mcp_golang.ToolResponse, error) {
			result, err := s.SymbolSource(workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}

// Outline lists symbols in a Go file or all files of a package directory.
func (s *GoASTService) Outline(workingDir string, args GoOutlineArgs) (string, error) {
	path, err := resolveWorkspacePath(workingDir, args.Path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("unable to access %s: %w", args.Path, err)
	}

	var files []string
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", fmt.Errorf("failed to list %s: %w", args.Path, err)
		}
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".go") {
				continue
			}
			if !args.IncludeTests && strings.HasSuffix(name, "_test.go") {
				continue
			}
			files = append(files, filepath.Join(path, name))
		}
		if len(files) == 0 {
			return fmt.Sprintf("No Go files found in %s", args.Path), nil
		}
	} else {
		files = []string{path}
	}

	var builder strings.Builder
	for _, file := range files {
		symbols, pkg, err := parseFileSymbols(workingDir, file)
		if err != nil {
			builder.WriteString(fmt.Sprintf("\n%s: parse error: %v\n", file, err))
			continue
		}
		rel, relErr := filepath.Rel(workingDir, file)
		if relErr != nil {
			rel = file
		}
		builder.WriteString(fmt.Sprintf("\n%s (package %s)\n", rel, pkg))
		for _, symbol := range symbols {
			builder.WriteString(fmt.Sprintf("  %4d: [%s] %s\n", symbol.Line, symbol.Kind, symbol.Signature))
		}
	}

	return truncateOutput(strings.TrimPrefix(builder.String(), "\n")), nil
}

// FindSymbol searches all project Go files for symbols matching a name.
func (s *GoASTService) FindSymbol(workingDir string, args GoFindSymbolArgs) (string, error) {
	if strings.TrimSpace(args.Name) == "" {
		return "", fmt.Errorf("argument 'name' is required")
	}

	maxResults := args.MaxResults
	if maxResults <= 0 {
		maxResults = 50
	}

	matcher := func(name string) bool {
		if args.Exact {
			return name == args.Name
		}
		return strings.Contains(strings.ToLower(name), strings.ToLower(args.Name))
	}

	var results []goSymbol
	err := walkGoFiles(workingDir, func(file string) error {
		symbols, _, err := parseFileSymbols(workingDir, file)
		if err != nil {
			return nil // skip unparseable files
		}
		for _, symbol := range symbols {
			if matcher(symbol.Name) {
				results = append(results, symbol)
				if len(results) >= maxResults {
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return fmt.Sprintf("No symbols matching %q found", args.Name), nil
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%d symbol(s) matching %q:\n", len(results), args.Name))
	for _, symbol := range results {
		builder.WriteString(fmt.Sprintf("• %s:%d [%s] %s\n", symbol.File, symbol.Line, symbol.Kind, symbol.Signature))
	}
	return truncateOutput(builder.String()), nil
}

// SymbolSource extracts the full source text of a named symbol.
func (s *GoASTService) SymbolSource(workingDir string, args GoSymbolSourceArgs) (string, error) {
	if strings.TrimSpace(args.Name) == "" {
		return "", fmt.Errorf("argument 'name' is required")
	}

	var files []string
	if strings.TrimSpace(args.Path) != "" {
		path, err := resolveWorkspacePath(workingDir, args.Path)
		if err != nil {
			return "", err
		}
		files = []string{path}
	} else {
		err := walkGoFiles(workingDir, func(file string) error {
			files = append(files, file)
			return nil
		})
		if err != nil {
			return "", err
		}
	}

	for _, file := range files {
		source, err := extractSymbolSource(workingDir, file, args.Name)
		if err != nil {
			continue
		}
		if source != "" {
			return truncateOutput(source), nil
		}
	}

	return "", fmt.Errorf("symbol %q not found", args.Name)
}

// walkGoFiles visits every non-generated Go file under the working directory.
func walkGoFiles(workingDir string, visit func(file string) error) error {
	return filepath.WalkDir(workingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		return visit(path)
	})
}

// parseFileSymbols parses a single Go file and extracts its top-level symbols.
func parseFileSymbols(workingDir, file string) ([]goSymbol, string, error) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file, nil, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, "", err
	}

	rel, relErr := filepath.Rel(workingDir, file)
	if relErr != nil {
		rel = file
	}

	var symbols []goSymbol
	for _, decl := range parsed.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			symbols = append(symbols, funcSymbol(fset, rel, d))
		case *ast.GenDecl:
			symbols = append(symbols, genDeclSymbols(fset, rel, d)...)
		}
	}

	return symbols, parsed.Name.Name, nil
}

func funcSymbol(fset *token.FileSet, file string, decl *ast.FuncDecl) goSymbol {
	kind := "func"
	signature := "func "
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		kind = "method"
		signature += "(" + types.ExprString(decl.Recv.List[0].Type) + ") "
	}
	signature += decl.Name.Name + strings.TrimPrefix(types.ExprString(decl.Type), "func")

	return goSymbol{
		Kind:      kind,
		Name:      decl.Name.Name,
		Signature: signature,
		File:      file,
		Line:      fset.Position(decl.Pos()).Line,
		EndLine:   fset.Position(decl.End()).Line,
		Doc:       decl.Doc.Text(),
	}
}

func genDeclSymbols(fset *token.FileSet, file string, decl *ast.GenDecl) []goSymbol {
	var symbols []goSymbol
	for _, spec := range decl.Specs {
		switch sp := spec.(type) {
		case *ast.TypeSpec:
			kind := "type"
			detail := ""
			switch t := sp.Type.(type) {
			case *ast.StructType:
				kind = "struct"
				if t.Fields != nil {
					detail = fmt.Sprintf(" {%d fields}", t.Fields.NumFields())
				}
			case *ast.InterfaceType:
				kind = "interface"
				if t.Methods != nil {
					detail = fmt.Sprintf(" {%d methods}", t.Methods.NumFields())
				}
			default:
				detail = " = " + types.ExprString(sp.Type)
			}
			symbols = append(symbols, goSymbol{
				Kind:      kind,
				Name:      sp.Name.Name,
				Signature: "type " + sp.Name.Name + detail,
				File:      file,
				Line:      fset.Position(sp.Pos()).Line,
				EndLine:   fset.Position(sp.End()).Line,
				Doc:       decl.Doc.Text(),
			})
		case *ast.ValueSpec:
			kind := strings.ToLower(decl.Tok.String())
			if kind != "const" && kind != "var" {
				continue
			}
			for _, name := range sp.Names {
				if name.Name == "_" {
					continue
				}
				signature := kind + " " + name.Name
				if sp.Type != nil {
					signature += " " + types.ExprString(sp.Type)
				}
				symbols = append(symbols, goSymbol{
					Kind:      kind,
					Name:      name.Name,
					Signature: signature,
					File:      file,
					Line:      fset.Position(name.Pos()).Line,
					EndLine:   fset.Position(sp.End()).Line,
					Doc:       decl.Doc.Text(),
				})
			}
		}
	}
	return symbols
}

// extractSymbolSource returns the exact source text of a named declaration,
// including its doc comment, or "" when the symbol is not in the file.
func extractSymbolSource(workingDir, file, name string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file, data, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return "", err
	}

	rel, relErr := filepath.Rel(workingDir, file)
	if relErr != nil {
		rel = file
	}

	extract := func(node ast.Node, doc *ast.CommentGroup, line int) string {
		start := node.Pos()
		if doc != nil {
			start = doc.Pos()
		}
		startOffset := fset.Position(start).Offset
		endOffset := fset.Position(node.End()).Offset
		header := fmt.Sprintf("// %s:%d\n", rel, line)
		return header + string(data[startOffset:endOffset])
	}

	for _, decl := range parsed.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.Name == name {
				return extract(d, d.Doc, fset.Position(d.Pos()).Line), nil
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch sp := spec.(type) {
				case *ast.TypeSpec:
					if sp.Name.Name == name {
						return extract(d, d.Doc, fset.Position(sp.Pos()).Line), nil
					}
				case *ast.ValueSpec:
					for _, ident := range sp.Names {
						if ident.Name == name {
							return extract(d, d.Doc, fset.Position(ident.Pos()).Line), nil
						}
					}
				}
			}
		}
	}

	return "", nil
}
