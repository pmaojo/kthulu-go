package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var makeServiceTestCmd = &cobra.Command{
	Use:   "make:service-test [name]",
	Short: "Genera una prueba básica para un servicio",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return scaffoldServiceTest(".", name)
	},
}

func init() {
	rootCmd.AddCommand(makeServiceTestCmd)
}

func scaffoldServiceTest(base, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("nombre de servicio inválido")
	}
	data := map[string]string{
		"ServiceName": exportName(name),
		"Package":     name,
	}
	dst := filepath.Join(base, "backend", "internal", name, "service_test.go")
	return writeTemplate("service_test.go.tmpl", dst, data, false)
}
