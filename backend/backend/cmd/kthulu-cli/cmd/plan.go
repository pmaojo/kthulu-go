package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/cli/depanalysis"
	planpkg "github.com/pmaojo/kthulu-go/backend/internal/cli/plan"
	"github.com/pmaojo/kthulu-go/backend/internal/cli/scanner"
	graphpkg "github.com/pmaojo/kthulu-go/backend/internal/graph"
	"gopkg.in/yaml.v3"
)

func newPlanCmd() *cobra.Command {
	var (
		graphOutput  bool
		outputFormat string
		validate     bool
	)

	cmd := &cobra.Command{
		Use:   "plan [dir]",
		Short: "Genera el plan de anotaciones de overrides y extends",
		Args:  cobra.MaximumNArgs(1),
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

	cmd.Flags().BoolVar(&graphOutput, "graph", false, "Genera y exporta grafo de validación")
	cmd.Flags().StringVar(&outputFormat, "format", "dot", "Formato de grafo: dot, json o yaml")
	cmd.Flags().BoolVar(&validate, "validate", false, "Valida grafo y reporta violaciones")

	return cmd
}

var planCmd = newPlanCmd()
