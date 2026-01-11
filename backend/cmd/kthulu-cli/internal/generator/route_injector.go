package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type RouteInjector struct {
	projectPath string
	fs          FileSystem
}

func NewRouteInjector(projectPath string) *RouteInjector {
	return &RouteInjector{
		projectPath: projectPath,
		fs:          RealFileSystem{},
	}
}

// InjectRoute adds the route to App.tsx if not present
func (r *RouteInjector) InjectRoute(moduleName, entityName string) error {
	appPath := filepath.Join(r.projectPath, "frontend", "src", "App.tsx")

	contentBytes, err := os.ReadFile(appPath)
	if err != nil {
		return fmt.Errorf("failed to read App.tsx: %w", err)
	}
	content := string(contentBytes)

	// Define import and route strings
	// import ProductPage from "@/modules/products/presentation/admin/Product";
	importLine := fmt.Sprintf("import %sPage from \"@/modules/%s/presentation/admin/%s\";", entityName, moduleName, entityName)

	// <Route path="/products/product" element={<ProductPage />} />
	// Heuristic for path: /module/entity
	routePath := fmt.Sprintf("/%s/%s", strings.ToLower(moduleName), strings.ToLower(entityName))
	routeLine := fmt.Sprintf("<Route path=\"%s\" element={<%sPage />} />", routePath, entityName)

	// Check if already exists
	if strings.Contains(content, importLine) && strings.Contains(content, routeLine) {
		return nil // Already injected
	}

	// Inject Import
	if !strings.Contains(content, importLine) {
		marker := "/* KTHULU:IMPORTS */"
		if strings.Contains(content, marker) {
			content = strings.Replace(content, marker, importLine+"\n"+marker, 1)
		} else {
			// Fallback: append after last import
			// Simple heuristic: find "import " and go to end of imports?
			// Markers are safer. If missing, maybe log warning.
			fmt.Println("Warning: Could not find /* KTHULU:IMPORTS */ marker in App.tsx")
		}
	}

	// Inject Route
	if !strings.Contains(content, routeLine) {
		marker := "/* KTHULU:ROUTES */"
		if strings.Contains(content, marker) {
			content = strings.Replace(content, marker, routeLine+"\n            "+marker, 1)
		} else {
			fmt.Println("Warning: Could not find /* KTHULU:ROUTES */ marker in App.tsx")
		}
	}

	return r.fs.WriteFile(appPath, []byte(content), 0644)
}
