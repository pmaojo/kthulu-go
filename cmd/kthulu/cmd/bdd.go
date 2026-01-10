package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var bddCmd = &cobra.Command{
	Use:   "bdd",
	Short: "Behavior Driven Development tools",
}

var bddFeaturesCmd = &cobra.Command{
	Use:   "features",
	Short: "List all feature files",
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, _ := os.Getwd()
		var features []string

		// Default search paths
		searchPaths := []string{"features", "backend/features"}

		found := false
		for _, searchPath := range searchPaths {
			fullPath := filepath.Join(workingDir, searchPath)
			if _, err := os.Stat(fullPath); err == nil {
				found = true
				err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && strings.HasSuffix(info.Name(), ".feature") {
						relPath, _ := filepath.Rel(workingDir, path)
						features = append(features, relPath)
					}
					return nil
				})
				if err != nil {
					return fmt.Errorf("failed to walk features dir: %w", err)
				}
			}
		}

		if !found {
			// Return empty list instead of error for API friendliness
			features = []string{}
		}

		// Output as line-delimited list (which the frontend expects splitting by \n)
		// Or JSON? The frontend `kthuluApi.ts` `listFeatures` mock implementation:
		// const paths = result.output[0].split('\n').filter(Boolean);
		// So plain text lines are expected.
		for _, f := range features {
			fmt.Println(f)
		}
		return nil
	},
}

var bddRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run BDD scenarios",
	RunE: func(cmd *cobra.Command, args []string) error {
		filter, _ := cmd.Flags().GetString("filter")
		workingDir, _ := os.Getwd()

		testPath := "./..."
		if _, err := os.Stat(filepath.Join(workingDir, "backend/features")); err == nil {
			testPath = "./backend/features/..."
		} else if _, err := os.Stat(filepath.Join(workingDir, "features")); err == nil {
			testPath = "./features/..."
		}

		cmdArgs := []string{"test", "-v", testPath}
		if filter != "" {
			cmdArgs = append(cmdArgs, "-args", filter)
		}

		// We just run and stream output to stdout
		c := exec.Command("go", cmdArgs...)
		c.Dir = workingDir
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		return c.Run()
	},
}

func init() {
	bddRunCmd.Flags().String("filter", "", "Filter scenarios to run")

	bddCmd.AddCommand(bddFeaturesCmd)
	bddCmd.AddCommand(bddRunCmd)

	rootCmd.AddCommand(bddCmd)
}
