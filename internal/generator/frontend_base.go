package generator

import (
	"fmt"
	"path"
)

// generateFrontendBase generates the base React files
func (g *TemplateGenerator) generateFrontendBase(structure *ProjectStructure) error {
	// Re-parse module features for App.tsx (capitalization handled in template usually, but we need the list)
	// g.config.Features has the list.
	
	// Create base templates map: targetPath -> templateName
	templates := map[string]string{
		"frontend/package.json":          "package.json.tmpl",
		"frontend/index.html":            "index.html.tmpl",
		"frontend/vite.config.ts":        "vite.config.ts.tmpl",
		"frontend/tsconfig.json":         "tsconfig.json.tmpl",
		"frontend/tsconfig.node.json":    "tsconfig.node.json.tmpl",
		"frontend/src/main.tsx":          "main.tsx.tmpl",
		"frontend/src/index.css":         "index.css.tmpl",
		"frontend/src/App.tsx":           "App.tsx.tmpl",
		"frontend/src/services/api.ts":   "api.ts.tmpl",
		"frontend/playwright.config.ts":  "e2e/playwright.config.ts.tmpl",
		"frontend/tests/e2e/auth.spec.ts": "e2e/auth.spec.ts.tmpl",
		// Admin UI
		"frontend/src/components/layouts/AdminLayout.tsx": "AdminLayout.tsx.tmpl",
		"frontend/src/config/navigation.ts":               "config/navigation.ts.tmpl",
	}

	// Convention over Configuration:
	// 'modules:' in schema = fullstack (get frontend)
	// 'features:' in schema = backend-only
	frontendModules := g.config.FrontendModules

	data := map[string]interface{}{
		"Name":     g.config.ProjectName,
		"Features": frontendModules,
		"Modules":  frontendModules,
	}

	for path, tmplName := range templates {
		// content, err := g.executeTemplate(tmplName, "scaffold/frontend/react/"+tmplName, data) 
		// Note: executeTemplate first arg is 'name' of template instance, second is path. 
		// I must check executeTemplate signature in the file.
		// Assuming g.executeTemplate(name, path, data)
		
		content, err := g.executeTemplate(tmplName, "scaffold/frontend/react/"+tmplName, data)
		if err != nil {
			return fmt.Errorf("failed to generate %s: %w", path, err)
		}

		structure.Files = append(structure.Files, GeneratedFile{
			Path:    path,
			Content: content,
		})
	}

	// Ensure src directory exists in structure
	structure.Directories = append(structure.Directories, 
		"frontend", 
		"frontend/src",
		"frontend/src/modules",
		"frontend/src/services",
		"frontend/src/components/layouts",
		"frontend/src/config",
		"frontend/tests/e2e",
	)

	return nil
}

// GenerateAdminModule generates the Admin Dashboard module in the frontend
func (g *TemplateGenerator) GenerateAdminModule(structure *ProjectStructure) error {
	fmt.Println("  🎨 Generating Admin Dashboard module...")

	moduleDir := "frontend/src/modules/admin/presentation"

	// Ensure directories exist
	structure.Directories = append(structure.Directories, moduleDir)

	// Map of file path -> template constant
	files := map[string]string{
		path.Join(moduleDir, "DashboardPage.tsx"): TmplAdminDashboard,
		path.Join(moduleDir, "UsersPage.tsx"):     TmplAdminUsers,
		path.Join(moduleDir, "SettingsPage.tsx"):  TmplAdminSettings,
	}

	for relPath, tmplName := range files {
		// Admin templates don't strictly need data for now, but we pass config just in case
		content, err := g.executeTemplate("admin_"+path.Base(relPath), tmplName, g.config)
		if err != nil {
			return fmt.Errorf("failed to generate %s: %w", relPath, err)
		}

		structure.Files = append(structure.Files, GeneratedFile{
			Path:    relPath,
			Content: content,
		})
	}

	return nil
}
