package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

const (
	backendTemplatesPath  = "cmd/kthulu/templates/backend/internal"
	usecasePath           = "cmd/kthulu/templates/backend/internal/usecase"
	httpPath              = "cmd/kthulu/templates/backend/internal/adapters/http"
	frontendTemplatesPath = "cmd/kthulu/templates/frontend/src/modules"
	registryPath          = "registry/modules"
)

type ModuleData struct {
	ID            string
	Title         string
	Description   string
	HasFrontend   bool
	EnvVars       []string
	Features      []string
	Stars         int
	Icon          string
	ManualContent string
}

func main() {
	fmt.Println("🚀 Starting Registry Generation...")

	modules, err := discoverModules()
	if err != nil {
		fmt.Printf("Error discovering modules: %v\n", err)
		os.Exit(1)
	}

	for _, mod := range modules {
		if err := generateModuleDoc(mod); err != nil {
			fmt.Printf("Error generating doc for %s: %v\n", mod.ID, err)
		} else {
			fmt.Printf("✅ Generated %s\n", mod.ID)
		}
	}

	fmt.Println("✨ Registry generation complete!")
}

func discoverModules() ([]ModuleData, error) {
	var modules []ModuleData

	files, err := os.ReadDir(backendTemplatesPath)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".go.tmpl") {
			continue
		}

		contentBytes, err := os.ReadFile(filepath.Join(backendTemplatesPath, f.Name()))
		if err != nil {
			continue
		}
		content := string(contentBytes)

		moduleID := extractModuleTag(content)
		if moduleID == "" {
			if strings.Contains(content, "fx.Options") {
				moduleID = strings.TrimSuffix(f.Name(), ".go.tmpl")
			} else {
				continue
			}
		}

		if moduleID == "registry" || moduleID == "routes" || moduleID == "module_set" {
			continue
		}

		fullContent := content + readRelatedFiles(moduleID)

		mod := ModuleData{
			ID:          moduleID,
			Title:       strings.Title(moduleID),
			Description: extractDescription(content),
			EnvVars:     extractEnvVars(fullContent),
			Features:    extractFeatures(fullContent),
			Stars:       0,
			Icon:        "Box",
		}

		if _, err := os.Stat(filepath.Join(frontendTemplatesPath, moduleID)); err == nil {
			mod.HasFrontend = true
		} else if _, err := os.Stat(filepath.Join(frontendTemplatesPath, moduleID+"s")); err == nil {
			mod.HasFrontend = true
		}

		modules = append(modules, mod)
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].ID < modules[j].ID
	})

	return modules, nil
}

func readRelatedFiles(moduleID string) string {
	var content strings.Builder

	// Check Usecase
	ucFile := filepath.Join(usecasePath, moduleID+".go.tmpl")
	if b, err := os.ReadFile(ucFile); err == nil {
		content.Write(b)
	}

	// Check HTTP Adapter
	httpFile := filepath.Join(httpPath, moduleID+".go.tmpl")
	if b, err := os.ReadFile(httpFile); err == nil {
		content.Write(b)
	}

	// Check internal folder if exists (e.g. backend/internal/auth/*.go.tmpl)
	internalDir := filepath.Join(backendTemplatesPath, moduleID)
	if entries, err := os.ReadDir(internalDir); err == nil {
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".go.tmpl") {
				if b, err := os.ReadFile(filepath.Join(internalDir, e.Name())); err == nil {
					content.Write(b)
				}
			}
		}
	}

	return content.String()
}

func extractModuleTag(content string) string {
	re := regexp.MustCompile(`@kthulu:module:(\w+)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractDescription(content string) string {
	var descLines []string
	varDefinitionLine := -1
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, "var") && strings.Contains(line, "Module = fx.Options") {
			varDefinitionLine = i
			break
		}
	}

	if varDefinitionLine > 0 {
		for i := varDefinitionLine - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, "//") {
				clean := strings.TrimSpace(strings.TrimPrefix(line, "//"))
				if !strings.HasPrefix(clean, "@") {
					descLines = append([]string{clean}, descLines...)
				}
			} else if line == "" {
				continue
			} else {
				break
			}
		}
	}

	if len(descLines) > 0 {
		return strings.Join(descLines, " ")
	}
	return "No description available."
}

func extractEnvVars(content string) []string {
	re := regexp.MustCompile(`os\.Getenv\("([^"]+)"\)`)
	matches := re.FindAllStringSubmatch(content, -1)

	unique := make(map[string]bool)
	var vars []string

	for _, m := range matches {
		if len(m) > 1 {
			v := m[1]
			if !unique[v] {
				unique[v] = true
				vars = append(vars, v)
			}
		}
	}

	reConfig := regexp.MustCompile(`config\.Get\w*\("([^"]+)"\)`)
	matchesConfig := reConfig.FindAllStringSubmatch(content, -1)
	for _, m := range matchesConfig {
		if len(m) > 1 {
			v := m[1]
			if !unique[v] {
				unique[v] = true
				vars = append(vars, v)
			}
		}
	}

	sort.Strings(vars)
	return vars
}

func extractFeatures(content string) []string {
	var features []string
	if strings.Contains(content, "Handler") || strings.Contains(content, "RegisterRoutes") {
		features = append(features, "HTTP API")
	}
	if strings.Contains(content, "UseCase") || strings.Contains(content, "Service") {
		features = append(features, "Domain Logic")
	}
	if strings.Contains(content, "Repository") {
		features = append(features, "Database Persistence")
	}
	if strings.Contains(content, "Middleware") {
		features = append(features, "Middleware")
	}

	unique := make(map[string]bool)
	var deduped []string
	for _, f := range features {
		if !unique[f] {
			unique[f] = true
			deduped = append(deduped, f)
		}
	}
	return deduped
}

const docTemplate = `---
title: "{{.Title}}"
description: "{{.Description}}"
type: "module"
author: "Kthulu Core"
stars: {{.Stars}}
icon: "{{.Icon}}"
---

# {{.Title}}

{{.Description}}

## Features

{{range .Features}}- {{.}}
{{else}}
- Auto-configured Fx Module
- Clean Architecture structure
{{end}}
{{if .HasFrontend}}
- **Frontend included**: React components and Admin UI integration.
{{end}}

## Configuration

The module is configured via environment variables:

| Variable | Description |
|----------|-------------|
{{range .EnvVars}}| ` + "`{{.}}`" + ` | Configuration for {{.}} |
{{else}}| - | No environment variables detected |
{{end}}

## Installation

Add this module to your project:

` + "```bash" + `
kthulu add module {{.ID}}
` + "```" + `

## Components

This module provides the following components to the application:

- **Backend**: Go module wired with Uber Fx.
{{if .HasFrontend}}- **Frontend**: React components in ` + "`src/modules/{{.ID}}`" + `.{{end}}
{{if .ManualContent}}
{{.ManualContent}}
{{end}}`

func generateModuleDoc(mod ModuleData) error {
	path := filepath.Join(registryPath, mod.ID)
	indexPath := filepath.Join(path, "index.md")

	// Try to read existing file to preserve metadata and manual content
	if contentBytes, err := os.ReadFile(indexPath); err == nil {
		content := string(contentBytes)

		// Preserve Stars
		reStars := regexp.MustCompile(`stars:\s*(\d+)`)
		if matches := reStars.FindStringSubmatch(content); len(matches) > 1 {
			if s, err := strconv.Atoi(matches[1]); err == nil {
				mod.Stars = s
			}
		}

		// Preserve Icon
		reIcon := regexp.MustCompile(`icon:\s*"([^"]+)"`)
		if matches := reIcon.FindStringSubmatch(content); len(matches) > 1 {
			mod.Icon = matches[1]
		}

		// Preserve Author
		reAuthor := regexp.MustCompile(`author:\s*"([^"]+)"`)
		if matches := reAuthor.FindStringSubmatch(content); len(matches) > 1 {
			// Actually we don't have Author in mod struct for output but let's assume Kthulu Core is fine
			// or we could add it. For now leaving default but keeping logic available.
		}

		// Preserve Manual Sections
		// Heuristic: Look for "## Usage", "## API Reference", or "## Troubleshooting"
		// and take everything from the earliest occurrence.

		indices := []int{}
		keywords := []string{"## Usage", "## API Reference", "## Troubleshooting"}

		for _, kw := range keywords {
			idx := strings.Index(content, kw)
			if idx != -1 {
				indices = append(indices, idx)
			}
		}

		if len(indices) > 0 {
			sort.Ints(indices)
			firstManualSection := indices[0]
			mod.ManualContent = content[firstManualSection:]
		}
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	f, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl, err := template.New("doc").Parse(docTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, mod)
}
