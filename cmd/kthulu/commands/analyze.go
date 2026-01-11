package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/pkg/kit"
	"github.com/pmaojo/kthulu-go/internal/adapters/cli/depanalysis"
	graphpkg "github.com/pmaojo/kthulu-go/internal/adapters/cli/graph"
	planpkg "github.com/pmaojo/kthulu-go/internal/adapters/cli/plan"
	"github.com/pmaojo/kthulu-go/internal/adapters/cli/scanner"
	"gopkg.in/yaml.v3"
)

func newAnalyzeCmd() *cobra.Command {
	var (
		graphOutput  bool
		outputFormat string
		validate     bool
	)

	cmd := &cobra.Command{
		Use:   "analyze [dir]",
		Short: "Analyze overrides and extends annotations in the project",
		Long: `Analyze the project to identify 'overrides' and 'extends' annotations,
verify dependency consistency, and optionally generate a validation graph.

Examples:
  kthulu analyze .
  kthulu analyze --graph --format=dot
  kthulu analyze --validate`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			base := "."
			if len(args) > 0 {
				base = args[0]
			}
			if err := depanalysis.Analyze(base); err != nil {
				return err
			}
			anns, err := scanner.Scan(base)
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
			if err := planpkg.Write(p, base); err != nil {
				return err
			}

			if graphOutput || validate {
				cfg, err := core.NewConfig()
				if err != nil {
					return err
				}
				g, err := BuildValidationGraph(cfg)
				if err != nil {
					return err
				}
				if validate {
					if err := graphpkg.ValidateGraph(g); err != nil {
						return err
					}
				}
				if graphOutput {
					var data []byte
					switch outputFormat {
					case "dot":
						data = []byte(g.ToDOT())
					case "json":
						data, err = g.ToJSON()
					case "yaml":
						data, err = yaml.Marshal(g)
					default:
						return fmt.Errorf("unsupported format: %s", outputFormat)
					}
					if err != nil {
						return err
					}
					path := fmt.Sprintf("/tmp/kthulu.graph.%s", outputFormat)
					if err := os.WriteFile(path, data, 0o644); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&graphOutput, "graph", false, "Generate and export validation graph")
	cmd.Flags().StringVar(&outputFormat, "format", "dot", "Graph format: dot, json or yaml")
	cmd.Flags().BoolVar(&validate, "validate", false, "Validate graph and report violations")

	return cmd
}

var analyzeCmd = newAnalyzeCmd()
