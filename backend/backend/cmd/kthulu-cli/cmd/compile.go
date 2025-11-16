package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	builder "github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/builder"
	planpkg "github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/plan"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/scanner"
)

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compila overrides y extensiones",
	RunE: func(cmd *cobra.Command, args []string) error {
		out, _ := cmd.Flags().GetString("out")
		watch, _ := cmd.Flags().GetBool("watch")

		run := func() error {
			anns, err := scanner.Scan(".")
			if err != nil {
				return err
			}
			constructs := make([]planpkg.Construct, len(anns))
			for i, a := range anns {
				constructs[i] = planpkg.Construct{
					ID:       fmt.Sprintf("%s:%s:%s", a.Mode, a.Module, a.Symbol),
					Path:     filepath.Join(a.Module, a.Symbol),
					Priority: a.Priority,
				}
			}
			p := planpkg.Build(constructs)
			if err := planpkg.Write(p, "."); err != nil {
				return err
			}
			return builder.Generate(filepath.Join(".kthulu", "plan.json"), out)
		}

		if watch {
			if err := run(); err != nil {
				return err
			}
			return watchDirs([]string{"overrides"}, []string{"extends"}, run)
		}
		return run()
	},
}

func init() {
	compileCmd.Flags().String("out", "build", "Directorio de salida")
	compileCmd.Flags().Bool("watch", false, "Observar cambios")
	rootCmd.AddCommand(compileCmd)
}

func watchDirs(overrides, extends []string, fn func() error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	addDir := func(dir string) error {
		return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return watcher.Add(path)
			}
			return nil
		})
	}
	for _, d := range append(overrides, extends...) {
		_ = addDir(d)
	}

	for {
		select {
		case e := <-watcher.Events:
			if e.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				if err := fn(); err != nil {
					fmt.Fprintln(os.Stderr, "compile error:", err)
				}
			}
			if e.Op&fsnotify.Create != 0 {
				info, err := os.Stat(e.Name)
				if err == nil && info.IsDir() {
					_ = addDir(e.Name)
				}
			}
		case err := <-watcher.Errors:
			return err
		}
	}
}
