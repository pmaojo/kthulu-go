package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MarketplaceItem represents a module, starter, or plugin
type MarketplaceItem struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"` // starter, module, plugin
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Stars       int               `json:"stars"`
	Tags        []string          `json:"tags"`
	Version     string            `json:"version"`
	Install     map[string]string `json:"install,omitempty"` // os -> url map for plugins
	Icon        string            `json:"icon,omitempty"`
}

type RegistryIndex struct {
	GeneratedAt time.Time         `json:"generated_at"`
	Items       []MarketplaceItem `json:"items"`
}

func main() {
	registryPath := flag.String("path", ".", "Path to the registry root")
	outputPath := flag.String("output", "index.json", "Path to the output index.json file")
	flag.Parse()

	absRegistryPath, err := filepath.Abs(*registryPath)
	if err != nil {
		fmt.Printf("Error resolving registry path: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Indexing registry at: %s\n", absRegistryPath)

	items, err := fetchMarketplaceItems(absRegistryPath)
	if err != nil {
		fmt.Printf("Error scanning registry: %v\n", err)
		os.Exit(1)
	}

	index := RegistryIndex{
		GeneratedAt: time.Now().UTC(),
		Items:       items,
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling index: %v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(*outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputPath, data, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated index with %d items at %s\n", len(items), *outputPath)
}

func fetchMarketplaceItems(rootPath string) ([]MarketplaceItem, error) {
	var items []MarketplaceItem
	categories := []string{"starters", "modules", "plugins"}

	for _, cat := range categories {
		catItems, err := scanCategoryItems(rootPath, cat)
		if err != nil {
			return nil, err
		}
		items = append(items, catItems...)
	}

	return items, nil
}

func scanCategoryItems(rootPath, category string) ([]MarketplaceItem, error) {
	catPath := filepath.Join(rootPath, category)

	// Check if category directory exists
	if _, err := os.Stat(catPath); os.IsNotExist(err) {
		// Try looking in public/category just in case, but usually for index generation we look at source
		// Assuming the structure is direct for the registry repo
		return nil, nil
	}

	entries, err := os.ReadDir(catPath)
	if err != nil {
		return nil, err
	}

	var items []MarketplaceItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		item := parseMarketplaceItem(catPath, entry.Name(), category)
		items = append(items, item)
	}
	return items, nil
}

func parseMarketplaceItem(catPath, dirName, category string) MarketplaceItem {
	itemPath := filepath.Join(catPath, dirName)
	var item MarketplaceItem

	// Try reading metadata.json
	metaData, err := os.ReadFile(filepath.Join(itemPath, "metadata.json"))
	if err == nil {
		if err := json.Unmarshal(metaData, &item); err != nil {
			fmt.Printf("Warning: Failed to parse metadata.json for %s/%s: %v\n", category, dirName, err)
		}
	}

	// Fallback/Defaults
	if item.ID == "" {
		item.ID = dirName
	}
	if item.Name == "" {
		item.Name = strings.Title(strings.ReplaceAll(dirName, "-", " "))
	}
	if item.Type == "" {
		item.Type = strings.TrimSuffix(category, "s") // singularize
	}
	if item.Description == "" {
		item.Description = fmt.Sprintf("Auto-discovered %s", item.Type)
	}
	if item.Author == "" {
		item.Author = "community"
	}

	return item
}
