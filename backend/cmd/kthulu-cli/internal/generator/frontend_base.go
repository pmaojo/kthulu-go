package generator

import (
	"fmt"
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
	}

	data := map[string]interface{}{
		"Name":     g.config.ProjectName,
		"Features": g.config.Features,
		"Modules":  g.config.Features, // Alias for template convenience
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
		"frontend/src/modules", // Already added by modules but good to be safe
	)

	return nil
}
