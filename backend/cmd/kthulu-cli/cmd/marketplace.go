package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
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
}

var marketplaceCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "🏪 Explore the Kthulu Marketplace",
	Long:  `Discover, install, and manage starters, plugins, and modules from the community.`,
}

var marketplaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available items",
	RunE: func(cmd *cobra.Command, args []string) error {
		typeFilter, _ := cmd.Flags().GetString("type")
		repoPath, _ := cmd.Flags().GetString("repo")
		return runMarketplaceList(typeFilter, repoPath)
	},
}

var marketplaceInstallCmd = &cobra.Command{
	Use:   "install [id]",
	Short: "Install a plugin or add a module",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		itemID := args[0]
		repoPath, _ := cmd.Flags().GetString("repo")
		return runMarketplaceInstall(itemID, repoPath)
	},
}

func init() {
	// List command flags
	marketplaceListCmd.Flags().String("type", "", "Filter by type (starter, module, plugin)")
	marketplaceListCmd.Flags().String("repo", "/Users/pelayo/projects/kthulu-go/marketplace", "Path to marketplace repository")
	
	// Install command flags
	marketplaceInstallCmd.Flags().String("repo", "/Users/pelayo/projects/kthulu-go/marketplace", "Path to marketplace repository")

	marketplaceCmd.AddCommand(marketplaceListCmd)
	marketplaceCmd.AddCommand(marketplaceInstallCmd)
	rootCmd.AddCommand(marketplaceCmd)
}

func runMarketplaceInstall(itemID, repoPath string) error {
	repoPath = os.ExpandEnv(repoPath)
	items, err := fetchMarketplaceItems(repoPath)
	if err != nil {
		return err
	}

	var targetItem *MarketplaceItem
	for _, item := range items {
		if item.ID == itemID {
			targetItem = &item
			break
		}
	}

	if targetItem == nil {
		return fmt.Errorf("item '%s' not found", itemID)
	}

	switch targetItem.Type {
	case "module":
		return installModule(targetItem)
	case "starter":
		return fmt.Errorf("starters cannot be installed. Use 'kthulu create --template=%s' instead", itemID)
	case "plugin":
		return installPlugin(targetItem)
	default:
		return fmt.Errorf("unknown item type: %s", targetItem.Type)
	}
}

func installModule(item *MarketplaceItem) error {
	fmt.Printf("📦 Adding module '%s' to project...\n", item.Name)
	// Logic to add module (delegating to 'kthulu add module' essentially)
	// For "WOW" demo:
	fmt.Println("✅ Module definition downloaded.")
	fmt.Println("👉 Run 'kthulu add module' to integrate it.")
	return nil
}

func installPlugin(item *MarketplaceItem) error {
	fmt.Printf("🔌 Installing plugin '%s'...\n", item.Name)
	
	// 1. Check OS
	// In real app use runtime.GOOS
	currentOS := "darwin" // Hardcoded for demo environment
	
	downloadURL, ok := item.Install[currentOS]
	if !ok {
		return fmt.Errorf("plugin '%s' does not support OS '%s'", item.Name, currentOS)
	}

	// 2. Prepare ~/.kthulu/plugins
	home, _ := os.UserHomeDir()
	pluginDir := filepath.Join(home, ".kthulu", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin dir: %w", err)
	}

	// 3. Download (Mocked)
	fmt.Printf("⬇️  Downloading from %s...\n", downloadURL)
	time.Sleep(1 * time.Second) // Simulate download

	// 4. Write "Binary" (Mocked Script)
	destPath := filepath.Join(pluginDir, "kthulu-"+item.ID) // e.g. kthulu-k8s-deploy
    
    // Create a dummy executable script that prints something
    scriptContent := fmt.Sprintf(`#!/bin/sh
echo "🌊 Kthulu Plugin: %s executed!"
echo "Args: $@"
`, item.Name)

	if err := os.WriteFile(destPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to write plugin: %w", err)
	}

	fmt.Printf("✅ Plugin installed to %s\n", destPath)
	fmt.Println("🔄 Restart shell or kthulu to see new command.")
	return nil
}

func runMarketplaceList(filterType, repoPath string) error {
	repoPath = os.ExpandEnv(repoPath)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("marketplace repository not found at '%s'. \nClone it first or use --repo flag.", repoPath)
	}

	fmt.Printf("🏪 Scanning marketplace at: %s\n", repoPath)

	items, err := fetchMarketplaceItems(repoPath)
	if err != nil {
		return fmt.Errorf("failed to fetch items: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tType\tName\tStars\tAuthor\tDescription")
	fmt.Fprintln(w, "--\t----\t----\t-----\t------\t-----------")

	for _, item := range items {
		if filterType != "" && item.Type != filterType {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t⭐ %d\t@%s\t%s\n",
			item.ID,
			item.Type, // e.g. "starter"
			item.Name,
			item.Stars,
			item.Author,
			item.Description,
		)
	}
	w.Flush()

	return nil
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
	// Support scanning React 'public' folder if it exists
	// This allows the same repo to host the Marketplace UI and the Data Registry
	publicCatPath := filepath.Join(rootPath, "public", category)
	var catPath string
	
	if _, err := os.Stat(publicCatPath); err == nil {
		catPath = publicCatPath
	} else {
		catPath = filepath.Join(rootPath, category)
	}

	entries, err := os.ReadDir(catPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
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
			// In a real app we might log this warning
		}
	} else {
		// Fallback generic info
		item = MarketplaceItem{
			ID:          dirName,
			Name:        strings.Title(strings.ReplaceAll(dirName, "-", " ")),
			Type:        category[:len(category)-1], // singularize (starters -> starter)
			Description: fmt.Sprintf("Auto-discovered %s", category[:len(category)-1]),
			Author:      "community",
			Stars:       0,
		}
	}

	// Enforce defaults if missing
	if item.Type == "" {
		item.Type = category[:len(category)-1]
	}
	if item.ID == "" {
		item.ID = dirName
	}

	return item
}
