package commands

import (
	"fmt"
	"os"
	"os/exec"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "📚 Generate API documentation",
	Long: `Generate Swagger/OpenAPI documentation for your project.
This command automatically checks for and installs the 'swag' tool if needed,
then runs 'swag init' to generate the documentation.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		diagrams, _ := cmd.Flags().GetBool("diagrams")
		return runDocCommand(cmd, args, diagrams)
	},
}

var (
	docDir         string
	docGeneralInfo string
)

func init() {
	docCmd.Flags().StringVar(&docDir, "dir", "", "Directory to search for main.go (default: project root)")
	docCmd.Flags().StringVar(&docGeneralInfo, "generalInfo", "", "API general info file (default: cmd/server/main.go)")
	docCmd.Flags().Bool("diagrams", false, "Generate Mermaid architecture diagrams")
	rootCmd.AddCommand(docCmd)
}

func runDocCommand(cmd *cobra.Command, args []string, generateDiagrams bool) error {
	fmt.Println("📚 Generating API documentation...")

	root, err := detectProjectRoot()
	if err != nil { return err }

	bin, err := ensureDocToolInstalled("swag", "github.com/swaggo/swag/cmd/swag@latest")
	if err != nil { return err }

	main := detectMainFile(root)
	
	if err := runSwaggerInit(bin, root, main); err != nil {
		return err
	}

	if generateDiagrams {
		fmt.Println("\n🎨 Generating Architecture Diagrams...")
		if err := generateMermaidDiagrams(root); err != nil {
			fmt.Printf("⚠️  Failed to generate diagrams: %v\n", err)
		}
	}

	fmt.Println("\n✅ Documentation generated successfully!")
	return nil
}

func detectProjectRoot() (string, error) {
	if docDir != "" {
		fmt.Printf("📂 Using specified directory: %s\n", docDir)
		return docDir, nil
	}
	root, err := findProjectRoot()
	if err != nil {
		fmt.Println("⚠️  Could not find go.mod, assuming current directory is project root.")
		return os.Getwd()
	}
	fmt.Printf("📂 Found project root: %s\n", root)
	return root, nil
}

func detectMainFile(root string) string {
	if docGeneralInfo != "" {
		fmt.Printf("📄 Using specified general info: %s\n", docGeneralInfo)
		return docGeneralInfo
	}
	candidates := []string{"cmd/server/main.go", "main.go", "cmd/kthulu-cli/main.go"}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(root, c)); err == nil {
			return c
		}
	}
	fmt.Println("⚠️  Could not find standard main file. Running swag init on root...")
	return "."
}

func runSwaggerInit(bin, root, main string) error {
	args := []string{"init", "--parseDependency", "--parseInternal"}
	if main != "." {
		args = append(args, "--generalInfo", main)
	}
	args = append(args, "--dir", root)

	run := exec.Command(bin, args...)
	run.Dir = root
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	return run.Run()
}

// generateMermaidDiagrams parses Go files and generates a class diagram
func generateMermaidDiagrams(root string) error {
	mermaidPath := filepath.Join(root, "docs", "architecture.mermaid")
	_ = os.MkdirAll(filepath.Join(root, "docs"), 0755)

	structs := make(map[string]string) // name -> package
	interfaces := []interfaceDef{}
	relationships := []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if skipFileForDiagram(path) { return nil }

		// Parse the file using go/parser
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil { return nil } // Skip files that can't be parsed

		pkg := node.Name.Name
		
		ast.Inspect(node, func(n ast.Node) bool {
			// Find type definitions
			ts, ok := n.(*ast.TypeSpec)
			if !ok { return true }

			name := ts.Name.Name

			// Handle Structs
			if st, ok := ts.Type.(*ast.StructType); ok {
				structs[name] = pkg
				// Check fields for relationships
				if st.Fields != nil {
					for _, field := range st.Fields.List {
						rel := extractRelationship(field.Type)
						if rel != "" && isCustomType(rel) {
							relationships = append(relationships, fmt.Sprintf("    %s --> %s", name, rel))
						}
					}
				}
			}

			// Handle Interfaces
			if _, ok := ts.Type.(*ast.InterfaceType); ok {
				interfaces = append(interfaces, interfaceDef{Name: name, Package: pkg})
			}

			return true
		})
		
		return nil
	})

	if err != nil { return err }

	var sb strings.Builder
	writeMermaidHeader(&sb)
	writeMermaidInterfaces(&sb, interfaces)
	writeMermaidClasses(&sb, structs)
	writeMermaidRelationships(&sb, relationships)

	return os.WriteFile(mermaidPath, []byte(sb.String()), 0644)
}

func extractRelationship(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return extractRelationship(t.X)
	case *ast.ArrayType:
		return extractRelationship(t.Elt)
	case *ast.SelectorExpr:
		return t.Sel.Name // Handling package.Type
	}
	return ""
}

func isCustomType(name string) bool {
	// Simple heuristic to ignore basic types
	basic := map[string]bool{
		"string": true, "int": true, "int64": true, "float64": true, 
		"bool": true, "error": true, "time.Time": true,
	}
	return !basic[name] && len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}

type interfaceDef struct {
	Name    string
	Package string
}

func skipFileForDiagram(path string) bool {
	return !strings.HasSuffix(path, ".go") || 
		strings.HasSuffix(path, "_test.go") || 
		strings.Contains(path, "vendor") || 
		strings.Contains(path, "mocks")
}

func writeMermaidHeader(sb *strings.Builder) {
	sb.WriteString("classDiagram\n")
	sb.WriteString("    direction TB\n\n")
}

func writeMermaidInterfaces(sb *strings.Builder, interfaces []interfaceDef) {
	for _, iface := range interfaces {
		sb.WriteString(fmt.Sprintf("    class %s {\n", iface.Name))
		sb.WriteString("        <<interface>>\n")
		sb.WriteString(fmt.Sprintf("        %s\n", iface.Package))
		sb.WriteString("    }\n")
	}
}

func writeMermaidClasses(sb *strings.Builder, structs map[string]string) {
	for name, pkg := range structs {
		sb.WriteString(fmt.Sprintf("    class %s {\n", name))
		sb.WriteString(fmt.Sprintf("        %s\n", pkg))
		sb.WriteString("    }\n")
	}
}

func writeMermaidRelationships(sb *strings.Builder, rels []string) {
	relMap := make(map[string]bool)
	for _, rel := range rels {
		if !relMap[rel] {
			sb.WriteString(rel + "\n")
			relMap[rel] = true
		}
	}
}


func ensureDocToolInstalled(name, installPath string) (string, error) {
	bin, err := exec.LookPath(name)
	if err == nil {
		return bin, nil
	}

	fmt.Printf("📦 Tool '%s' not found. Installing...\n", name)
	
	// Try to find in GOPATH/bin first
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = filepath.Join(os.Getenv("HOME"), "go")
	}
	
	binPath := filepath.Join(goPath, "bin", name)
	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	if err := runGoInstall(installPath); err != nil {
		return "", err
	}

	return binPath, nil
}

func runGoInstall(path string) error {
	install := exec.Command("go", "install", path)
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	return install.Run()
}

func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("go.mod not found")
		}
		wd = parent
	}
}
