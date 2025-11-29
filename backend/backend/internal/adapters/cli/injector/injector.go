package injector

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

// InjectFunction parses the original source code, appends the new function code,
// and returns the formatted result.
// originalSource: The content of the existing Go file.
// newFunctionSource: The content of the new function to add.
func InjectFunction(originalSource string, newFunctionSource string) (string, error) {
	fset := token.NewFileSet()

	// Parse the original file
	file, err := parser.ParseFile(fset, "original.go", originalSource, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse original source: %w", err)
	}

	// Parse the new function.
	// We wrap it in a dummy package because ParseFile expects a complete file.
	dummySource := fmt.Sprintf("package dummy\n\n%s", newFunctionSource)
	newFile, err := parser.ParseFile(fset, "new.go", dummySource, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse new function source: %w", err)
	}

	// Find the function declaration in the new file
	var newFuncDecl *ast.FuncDecl
	for _, decl := range newFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			newFuncDecl = funcDecl
			break
		}
	}

	if newFuncDecl == nil {
		return "", fmt.Errorf("no function declaration found in new function source")
	}

	// Check if a function with the same name and receiver already exists
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == newFuncDecl.Name.Name {
				// Check receivers
				if areReceiversSame(funcDecl.Recv, newFuncDecl.Recv) {
					return "", fmt.Errorf("function %s already exists", newFuncDecl.Name.Name)
				}
			}
		}
	}

	// Append the new function to the declarations
	file.Decls = append(file.Decls, newFuncDecl)

	// Print the modified AST to a buffer
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return "", fmt.Errorf("failed to format modified source: %w", err)
	}

	// Ensure proper spacing and formatting
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to format source with go/format: %w", err)
	}

	return string(formatted), nil
}

// areReceiversSame checks if two method receivers are effectively the same.
func areReceiversSame(r1, r2 *ast.FieldList) bool {
	// Both nil -> both are functions -> same
	if r1 == nil && r2 == nil {
		return true
	}
	// One nil, one not -> different
	if r1 == nil || r2 == nil {
		return false
	}

	// Compare the type expressions roughly by printing them.
	// This is a simplification but works for most cases (e.g., (s *Service) vs (s Service)).
	// Ideally we would deep compare the AST nodes, but this suffices for a collision check.
	// Note: We ignore the parameter names in the receiver (e.g., 's' vs 'x') and only care about the type.

	if len(r1.List) != len(r2.List) {
		return false
	}

	// Assuming single receiver argument which is standard Go
	if len(r1.List) > 0 {
		type1 := formatNode(r1.List[0].Type)
		type2 := formatNode(r2.List[0].Type)
		return type1 == type2
	}

	return true
}

func formatNode(node ast.Node) string {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	if err := format.Node(&buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}
