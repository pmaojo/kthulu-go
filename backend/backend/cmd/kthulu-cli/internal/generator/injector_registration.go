package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// InjectModuleRegistration finds cmd/server/main.go and registers the new module.
func InjectModuleRegistration(projectRoot, moduleName, projectModule string) error {
	mainPath := filepath.Join(projectRoot, "cmd", "server", "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		// Fallback to service if server not found (backward compatibility)
		mainPath = filepath.Join(projectRoot, "cmd", "service", "main.go")
		if _, err := os.Stat(mainPath); os.IsNotExist(err) {
			return fmt.Errorf("could not find main.go in cmd/server or cmd/service")
		}
	}

	content, err := os.ReadFile(mainPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", content, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse main.go: %w", err)
	}

	// Sanitize module name for package identifier
	safeModuleName := strings.ToLower(strings.ReplaceAll(moduleName, "-", ""))

	// 1. Add Import
	moduleImportPath := fmt.Sprintf("%s/internal/modules/%s", projectModule, moduleName)
	// We need to use a named import if the package name (sanitized) differs from the last path element
	// or just to be safe if moduleName has hyphens.
	// However, astutil.AddNamedImport is not available in all versions, or we can just rely on AddImport
	// if the package declaration in the module file matches the sanitized name.
	// The scaffold template uses `package {{.Name}}` which is usually the module name.
	// If the user provided `my-module`, the package name should be `mymodule`.

	// Let's assume the package name is safeModuleName.
	if safeModuleName != moduleName {
		astutil.AddNamedImport(fset, file, safeModuleName, moduleImportPath)
	} else {
		astutil.AddImport(fset, file, moduleImportPath)
	}

	// 2. Find fx.New and inject safeModuleName.Providers()
	foundFxNew := false
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if function call is fx.New
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "fx" && sel.Sel.Name == "New" {
				foundFxNew = true

				// Create the new argument: <safeModuleName>.Providers()
				newArg := &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: safeModuleName},
						Sel: &ast.Ident{Name: "Providers"},
					},
					Args: []ast.Expr{},
				}

				// Append to fx.New arguments
				// We want to insert it before fx.Invoke if possible, or just append
				call.Args = insertArgBeforeInvoke(call.Args, newArg)

				return false // Stop inspecting
			}
		}
		return true
	})

	if !foundFxNew {
		return fmt.Errorf("could not find fx.New() call in main.go")
	}

	// Write back
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return fmt.Errorf("failed to format modified main.go: %w", err)
	}

	// Format source to be safe
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to run gofmt: %w", err)
	}

	return os.WriteFile(mainPath, formatted, 0644)
}

func insertArgBeforeInvoke(args []ast.Expr, newArg ast.Expr) []ast.Expr {
	// Try to find the first fx.Invoke to insert before it
	insertIdx := -1
	for i, arg := range args {
		if isFxInvoke(arg) {
			insertIdx = i
			break
		}
	}

	if insertIdx == -1 {
		// Just append if no Invoke found
		return append(args, newArg)
	}

	// Insert at index
	newArgs := make([]ast.Expr, 0, len(args)+1)
	newArgs = append(newArgs, args[:insertIdx]...)
	newArgs = append(newArgs, newArg)
	newArgs = append(newArgs, args[insertIdx:]...)
	return newArgs
}

func isFxInvoke(arg ast.Expr) bool {
	call, ok := arg.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return pkg.Name == "fx" && sel.Sel.Name == "Invoke"
}
