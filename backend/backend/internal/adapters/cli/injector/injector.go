package injector

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
)

// InjectFunction parses the original source code, appends the new function code,
// adds necessary imports, and returns the formatted result.
// originalSource: The content of the existing Go file.
// newFunctionSource: The content of the new function to add.
// imports: A list of imports to add (e.g., "time", "net/http").
func InjectFunction(originalSource string, newFunctionSource string, imports []string) (string, error) {
	fset := token.NewFileSet()

	// Parse the original file
	file, err := parser.ParseFile(fset, "original.go", originalSource, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse original source: %w", err)
	}

	// Parse the new function.
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

	// Check collision
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == newFuncDecl.Name.Name {
				if areReceiversSame(funcDecl.Recv, newFuncDecl.Recv) {
					return "", fmt.Errorf("function %s already exists", newFuncDecl.Name.Name)
				}
			}
		}
	}

	// Add imports using astutil
	for _, imp := range imports {
		astutil.AddImport(fset, file, imp)
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

// InjectStructTag updates the tags for a specific field in a struct.
func InjectStructTag(originalSource string, structName string, fieldName string, newTags string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "original.go", originalSource, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse original source: %w", err)
	}

	var structType *ast.StructType
	foundStruct := false

	// Find the struct
	ast.Inspect(file, func(n ast.Node) bool {
		if t, ok := n.(*ast.TypeSpec); ok {
			if t.Name.Name == structName {
				if st, ok := t.Type.(*ast.StructType); ok {
					structType = st
					foundStruct = true
					return false
				}
			}
		}
		return true
	})

	if !foundStruct {
		return "", fmt.Errorf("struct %s not found", structName)
	}

	foundField := false
	// Find the field
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if name.Name == fieldName {
				// Update the tag
				if field.Tag == nil {
					field.Tag = &ast.BasicLit{}
				}
				field.Tag.Kind = token.STRING
				field.Tag.Value = "`" + newTags + "`"
				foundField = true
				break
			}
		}
		if foundField {
			break
		}
	}

	if !foundField {
		return "", fmt.Errorf("field %s not found in struct %s", fieldName, structName)
	}

	// Format
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return "", fmt.Errorf("failed to format modified source: %w", err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to format source with go/format: %w", err)
	}

	return string(formatted), nil
}

// areReceiversSame checks if two method receivers are effectively the same.
func areReceiversSame(r1, r2 *ast.FieldList) bool {
	if r1 == nil && r2 == nil {
		return true
	}
	if r1 == nil || r2 == nil {
		return false
	}

	if len(r1.List) != len(r2.List) {
		return false
	}

	if len(r1.List) > 0 {
		type1 := getBaseTypeName(r1.List[0].Type)
		type2 := getBaseTypeName(r2.List[0].Type)
		return type1 == type2
	}

	return true
}

func getBaseTypeName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		return formatNode(star.X)
	}
	return formatNode(expr)
}

func formatNode(node ast.Node) string {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	if err := format.Node(&buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}

// InjectStructField adds a new field to an existing struct.
func InjectStructField(originalSource string, structName string, fieldName string, fieldType string, tags string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "original.go", originalSource, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse original source: %w", err)
	}

	var structType *ast.StructType
	found := false

	// Find the struct
	ast.Inspect(file, func(n ast.Node) bool {
		if t, ok := n.(*ast.TypeSpec); ok {
			if t.Name.Name == structName {
				if st, ok := t.Type.(*ast.StructType); ok {
					structType = st
					found = true
					return false
				}
			}
		}
		return true
	})

	if !found {
		return "", fmt.Errorf("struct %s not found", structName)
	}

	// Check if field already exists
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if name.Name == fieldName {
				return "", fmt.Errorf("field %s already exists in struct %s", fieldName, structName)
			}
		}
	}

	// Create new field
	// We parse a dummy struct to get the AST for the field
	dummyStruct := fmt.Sprintf("package dummy\ntype D struct { %s %s `%s` }", fieldName, fieldType, tags)
	dummyFile, err := parser.ParseFile(fset, "dummy.go", dummyStruct, 0)
	if err != nil {
		return "", fmt.Errorf("failed to parse field definition: %w", err)
	}

	var newField *ast.Field
	ast.Inspect(dummyFile, func(n ast.Node) bool {
		if t, ok := n.(*ast.TypeSpec); ok {
			if st, ok := t.Type.(*ast.StructType); ok {
				if len(st.Fields.List) > 0 {
					newField = st.Fields.List[0]
				}
			}
		}
		return true
	})

	if newField == nil {
		return "", fmt.Errorf("failed to create new field AST")
	}

	// Append field
	structType.Fields.List = append(structType.Fields.List, newField)

	// Format
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return "", fmt.Errorf("failed to format modified source: %w", err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to format source with go/format: %w", err)
	}

	return string(formatted), nil
}
