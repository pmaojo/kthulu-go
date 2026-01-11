package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/generator"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/parser"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Manage the admin interface",
}

var watch bool

var generateAdminCmd = &cobra.Command{
	Use:   "generate [module]",
	Short: "Generate admin UI for a module",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRoot, err := os.Getwd()
		if err != nil {
			return err
		}
		// Basic check for root
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); os.IsNotExist(err) {
			parent := filepath.Dir(projectRoot)
			if _, err := os.Stat(filepath.Join(parent, "go.mod")); err == nil {
				projectRoot = parent
			}
		}

		targetModule := ""
		if len(args) > 0 {
			targetModule = args[0]
		}

		// Handle path resolution whether we are in root or backend
		domainPath := filepath.Join(projectRoot, "backend", "internal", "domain")
		if _, err := os.Stat(domainPath); os.IsNotExist(err) {
			// Try without "backend" prefix
			domainPath = filepath.Join(projectRoot, "internal", "domain")
		}

		runGeneration := func() error {
			fmt.Println("Running generation...")

			// 0. Ensure Frontend Exists
			scaffolder := generator.NewFrontendScaffolder(projectRoot)
			if err := scaffolder.EnsureFrontend(); err != nil {
				return fmt.Errorf("failed to scaffold frontend: %w", err)
			}

			p := parser.NewEntityParser(projectRoot)

			files, err := os.ReadDir(domainPath)
			if err != nil {
				return fmt.Errorf("failed to read domain directory: %w", err)
			}

			gen := generator.NewAdminGenerator(projectRoot)
			injector := generator.NewRouteInjector(projectRoot)

			for _, file := range files {
				if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
					continue
				}

				filePath := filepath.Join(domainPath, file.Name())
				entities, err := p.ParseFile(filePath)
				if err != nil {
					fmt.Printf("Warning: failed to parse %s: %v\n", file.Name(), err)
					continue
				}

				for _, entity := range entities {
					if entity.Module == "" {
						// Simplistic module inference
						entity.Module = strings.TrimSuffix(file.Name(), ".go") + "s"
					}

					if targetModule != "" && entity.Module != targetModule && !strings.Contains(entity.Module, targetModule) {
						continue
					}

					fmt.Printf("Generating Admin UI for %s (Module: %s)...\n", entity.Name, entity.Module)

					uiDef := generator.MapEntityToUI(entity)
					if err := gen.GenerateAdminModule(uiDef); err != nil {
						return fmt.Errorf("failed to generate admin for %s: %w", entity.Name, err)
					}

					// Inject Routes
					if err := injector.InjectRoute(entity.Module, entity.Name); err != nil {
						fmt.Printf("Warning: Failed to inject route for %s: %v\n", entity.Name, err)
					}
				}
			}
			return nil
		}

		// Initial Run
		if err := runGeneration(); err != nil {
			return err
		}

		if watch {
			fmt.Println("Watching for changes in", domainPath)
			return runWatcher(domainPath, runGeneration)
		}

		return nil
	},
}

func init() {
	generateAdminCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes and regenerate")
	adminCmd.AddCommand(generateAdminCmd)
	rootCmd.AddCommand(adminCmd)
}

func runWatcher(path string, action func() error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("File modified:", event.Name)
					// Debounce or just run
					time.Sleep(100 * time.Millisecond) // Simple debounce
					if err := action(); err != nil {
						fmt.Println("Error generating:", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Watcher error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		return err
	}
	<-done
	return nil
}
