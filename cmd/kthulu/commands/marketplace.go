package commands

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/registry"
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
	Template    string            `json:"template,omitempty"` // for starters scaffolded via create --template
	Install     map[string]string `json:"install,omitempty"`  // os -> url map for plugins
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
	Use:   "install [id] [project-name]",
	Short: "Install an item: scaffold starters, add modules, install plugins",
	Long: `Install a marketplace item.

Starters scaffold a complete project: the starter's blueprint is copied next
to the new project and 'kthulu create --from-plan' runs automatically.

  kthulu marketplace install ecommerce-pro my-shop`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		itemID := args[0]
		projectName := ""
		if len(args) > 1 {
			projectName = args[1]
		}
		repoPath, _ := cmd.Flags().GetString("repo")
		return runMarketplaceInstall(itemID, projectName, repoPath)
	},
}

func init() {
	// List command flags
	marketplaceListCmd.Flags().String("type", "", "Filter by type (starter, module, plugin)")
	marketplaceListCmd.Flags().String("repo", "", "Path to a marketplace registry (default: ./registry if present, else the registry embedded in the binary)")

	// Install command flags
	marketplaceInstallCmd.Flags().String("repo", "", "Path to a marketplace registry (default: ./registry if present, else the registry embedded in the binary)")

	marketplaceCmd.AddCommand(marketplaceListCmd)
	marketplaceCmd.AddCommand(marketplaceInstallCmd)
	rootCmd.AddCommand(marketplaceCmd)
}

func runMarketplaceInstall(itemID, projectName, repoPath string) error {
	source, label, err := resolveRegistryFS(repoPath)
	if err != nil {
		return err
	}
	items, err := fetchMarketplaceItems(source)
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
		return fmt.Errorf("item '%s' not found in %s", itemID, label)
	}

	switch targetItem.Type {
	case "module":
		return installModule(targetItem)
	case "starter":
		return installStarter(targetItem, source, projectName)
	case "plugin":
		return installPlugin(targetItem)
	default:
		return fmt.Errorf("unknown item type: %s", targetItem.Type)
	}
}

// resolveRegistryFS picks the registry source: an explicit --repo path, a
// local ./registry checkout, or the registry embedded in the binary.
func resolveRegistryFS(repoPath string) (fs.FS, string, error) {
	if repoPath != "" {
		repoPath = os.ExpandEnv(repoPath)
		if _, err := os.Stat(repoPath); err != nil {
			return nil, "", fmt.Errorf("marketplace registry not found at '%s'", repoPath)
		}
		return os.DirFS(repoPath), repoPath, nil
	}
	if info, err := os.Stat("registry"); err == nil && info.IsDir() {
		return os.DirFS("registry"), "./registry", nil
	}
	return registry.Files, "the embedded registry", nil
}

// installStarter scaffolds a project from a starter: it copies the starter's
// blueprint next to the project and runs 'kthulu create' on it.
func installStarter(item *MarketplaceItem, source fs.FS, projectName string) error {
	if projectName == "" {
		projectName = strings.ReplaceAll(item.ID, "-", "")
	}
	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("directory %q already exists — pass a different project name: kthulu marketplace install %s <name>", projectName, item.ID)
	}

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve kthulu binary: %w", err)
	}

	fmt.Printf("🚀 Scaffolding %q from starter '%s'...\n", projectName, item.Name)

	var createArgs []string
	runHint := "go run ./cmd/server"
	if plan, planErr := findStarterPlan(source, item.ID); planErr == nil {
		planFile := projectName + "-plan.yaml"
		if writeErr := os.WriteFile(planFile, plan, 0o644); writeErr != nil {
			return fmt.Errorf("write plan file: %w", writeErr)
		}
		fmt.Printf("   📄 Blueprint copied to %s\n", planFile)
		createArgs = []string{"create", projectName, "--from-plan", planFile}
	} else if item.Template != "" {
		createArgs = []string{"create", projectName, "--template", item.Template}
		runHint = "go run ./cmd/" + projectName
	} else {
		return fmt.Errorf("starter '%s' has no blueprint (plan.yaml/blueprint.yaml) and no template: %w", item.ID, planErr)
	}

	cmd := exec.Command(self, createArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create failed: %w", err)
	}

	fmt.Printf("\n✅ Starter '%s' installed. Next:\n   cd %s && %s\n", item.Name, projectName, runHint)
	return nil
}

// findStarterPlan loads the starter blueprint (plan.yaml, falling back to
// blueprint.yaml) from the registry source.
func findStarterPlan(source fs.FS, starterID string) ([]byte, error) {
	var lastErr error
	for _, name := range []string{"plan.yaml", "blueprint.yaml"} {
		data, err := fs.ReadFile(source, path.Join("starters", starterID, name))
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func installModule(item *MarketplaceItem) error {
	if _, err := os.Stat("go.mod"); err != nil {
		return fmt.Errorf("not inside a project (go.mod not found) — run this from your project directory")
	}
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve kthulu binary: %w", err)
	}
	fmt.Printf("📦 Adding module '%s' to the current project...\n", item.Name)
	cmd := exec.Command(self, "add", "module", item.ID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
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
	source, label, err := resolveRegistryFS(repoPath)
	if err != nil {
		return err
	}

	fmt.Printf("🏪 Scanning marketplace: %s\n", label)

	items, err := fetchMarketplaceItems(source)
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

func fetchMarketplaceItems(source fs.FS) ([]MarketplaceItem, error) {
	var items []MarketplaceItem
	categories := []string{"starters", "modules", "plugins"}

	for _, cat := range categories {
		catItems, err := scanCategoryItems(source, cat)
		if err != nil {
			return nil, err
		}
		items = append(items, catItems...)
	}

	return items, nil
}

func scanCategoryItems(source fs.FS, category string) ([]MarketplaceItem, error) {
	// Support scanning a React 'public' folder if it exists, so the same
	// repo can host the Marketplace UI and the Data Registry.
	catPath := category
	if _, err := fs.Stat(source, path.Join("public", category)); err == nil {
		catPath = path.Join("public", category)
	}

	entries, err := fs.ReadDir(source, catPath)
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

		item := parseMarketplaceItem(source, catPath, entry.Name(), category)
		items = append(items, item)
	}
	return items, nil
}

func parseMarketplaceItem(source fs.FS, catPath, dirName, category string) MarketplaceItem {
	var item MarketplaceItem

	// Try reading metadata.json
	metaData, err := fs.ReadFile(source, path.Join(catPath, dirName, "metadata.json"))
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
